package electrumx

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"sync"
	"time"

	"ginproject/entity/config"
	"ginproject/middleware/log"
)

// ConnPool ElectrumX连接池
type ConnPool struct {
	mu            sync.Mutex
	config        *config.ElectrumXConfig
	client        *ElectrumXClient
	conns         chan net.Conn
	maxIdleConns  int
	maxOpenConns  int
	connTimeout   time.Duration
	idleTimeout   time.Duration
	createdConns  int
	connErr       error
	lastConnErr   time.Time
	closed        bool
	cleanerCtx    context.Context
	cleanerCancel context.CancelFunc
}

// PoolConfig 连接池配置
type PoolConfig struct {
	MaxIdleConns int
	MaxOpenConns int
	IdleTimeout  time.Duration
}

// 常量定义
const (
	defaultIdleTimeout  = 10 * time.Minute
	defaultMaxIdleConns = 10
	defaultMaxOpenConns = 20
	connRetryDelay      = 5 * time.Second
)

var (
	// ErrPoolClosed 池已关闭错误
	ErrPoolClosed = errors.New("连接池已关闭")
	// ErrNoFreeConn 无可用连接错误
	ErrNoFreeConn = errors.New("无可用连接")
	// ErrConnTimeout 连接超时错误
	ErrConnTimeout = errors.New("获取连接超时")
)

// NewConnPool 创建一个新的连接池
func NewConnPool(client *ElectrumXClient, poolConfig *PoolConfig) (*ConnPool, error) {
	if client == nil {
		return nil, errors.New("ElectrumX客户端不能为空")
	}

	// 获取 ElectrumX 配置
	electrumXConfig := config.GetConfig().GetElectrumXConfig()
	if electrumXConfig == nil {
		return nil, fmt.Errorf("获取ElectrumX配置失败")
	}

	// 设置默认值
	maxIdleConns := defaultMaxIdleConns
	maxOpenConns := defaultMaxOpenConns
	idleTimeout := defaultIdleTimeout

	// 优先使用配置中的值
	if electrumXConfig.MaxIdleConns > 0 {
		maxIdleConns = electrumXConfig.MaxIdleConns
	}
	if electrumXConfig.MaxOpenConns > 0 {
		maxOpenConns = electrumXConfig.MaxOpenConns
	}

	// 如果提供了池配置，则覆盖默认值
	if poolConfig != nil {
		if poolConfig.MaxIdleConns > 0 {
			maxIdleConns = poolConfig.MaxIdleConns
		}
		if poolConfig.MaxOpenConns > 0 {
			maxOpenConns = poolConfig.MaxOpenConns
		}
		if poolConfig.IdleTimeout > 0 {
			idleTimeout = poolConfig.IdleTimeout
		}
	}

	// 确保参数合理
	if maxIdleConns > maxOpenConns {
		maxIdleConns = maxOpenConns
	}

	ctx, cancel := context.WithCancel(context.Background())

	pool := &ConnPool{
		client:        client,
		config:        electrumXConfig,
		conns:         make(chan net.Conn, maxIdleConns),
		maxIdleConns:  maxIdleConns,
		maxOpenConns:  maxOpenConns,
		connTimeout:   time.Duration(electrumXConfig.Timeout) * time.Second,
		idleTimeout:   idleTimeout,
		cleanerCtx:    ctx,
		cleanerCancel: cancel,
	}

	// 启动空闲连接清理协程
	go pool.connectionCleaner()

	log.Info("ElectrumX连接池已创建, 最大空闲连接:", maxIdleConns,
		", 最大打开连接:", maxOpenConns,
		", 空闲超时:", idleTimeout)
	return pool, nil
}

// GetConn 从池中获取一个连接
func (p *ConnPool) GetConn(ctx context.Context) (net.Conn, error) {
	p.mu.Lock()
	if p.closed {
		p.mu.Unlock()
		return nil, ErrPoolClosed
	}

	// 首先尝试从空闲池获取连接
	select {
	case conn := <-p.conns:
		p.mu.Unlock()
		// 检查连接是否有效
		if !p.validateConn(conn) {
			// 无效连接，创建新连接
			return p.createConn(ctx)
		}
		log.Debug("从连接池获取连接成功")
		return conn, nil
	default:
		// 空闲池中没有连接
	}

	// 检查是否可以创建新连接
	if p.createdConns < p.maxOpenConns {
		p.createdConns++
		p.mu.Unlock()
		return p.createConn(ctx)
	}

	// 已达到最大连接数，等待连接归还
	if ctx.Done() == nil {
		// 如果没有上下文超时，设置默认超时
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, p.connTimeout)
		defer cancel()
	}

	p.mu.Unlock()

	select {
	case conn := <-p.conns:
		if !p.validateConn(conn) {
			// 无效连接，尝试重新获取
			return p.GetConn(ctx)
		}
		return conn, nil
	case <-ctx.Done():
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return nil, ErrConnTimeout
		}
		return nil, ctx.Err()
	}
}

// createConn 创建新连接
func (p *ConnPool) createConn(ctx context.Context) (net.Conn, error) {
	// 如果最近连接有错误，等待一段时间再重试
	p.mu.Lock()
	if p.connErr != nil && time.Since(p.lastConnErr) < connRetryDelay {
		err := p.connErr
		p.mu.Unlock()
		return nil, err
	}
	p.mu.Unlock()

	// 设置超时上下文
	if ctx.Done() == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, p.connTimeout)
		defer cancel()
	}

	// 创建连接
	conn, err := p.client.Connect()
	if err != nil {
		p.mu.Lock()
		p.connErr = err
		p.lastConnErr = time.Now()
		p.createdConns--
		p.mu.Unlock()
		log.Error("创建ElectrumX连接失败:", err)
		return nil, err
	}

	log.Debug("创建新的ElectrumX连接成功")
	return conn, nil
}

// PutConn 将连接放回池中
func (p *ConnPool) PutConn(conn net.Conn) {
	if conn == nil {
		return
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		conn.Close()
		return
	}

	// 尝试将连接放入空闲池
	select {
	case p.conns <- conn:
		// 成功放入空闲池
		return
	default:
		// 空闲池已满，关闭连接
		p.createdConns--
		conn.Close()
	}
}

// validateConn 验证连接是否有效
func (p *ConnPool) validateConn(conn net.Conn) bool {
	if conn == nil {
		return false
	}

	// 设置1秒超时进行ping测试
	conn.SetDeadline(time.Now().Add(time.Second))
	defer conn.SetDeadline(time.Time{}) // 重置超时

	// 构建ping请求
	pingReq := RPCRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "server.ping",
		Params:  []interface{}{},
	}

	pingBytes, err := json.Marshal(pingReq)
	if err != nil {
		log.Error("序列化ping请求失败:", err)
		return false
	}
	pingBytes = append(pingBytes, '\n')

	// 发送ping请求
	if _, err = conn.Write(pingBytes); err != nil {
		log.Error("验证连接发送ping失败:", err)
		return false
	}

	// 读取响应
	reader := bufio.NewReader(conn)
	_, err = reader.ReadBytes('\n')
	if err != nil {
		log.Error("验证连接读取ping响应失败:", err)
		return false
	}

	return true
}

// Close 关闭连接池
func (p *ConnPool) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return ErrPoolClosed
	}

	// 停止清理协程
	p.cleanerCancel()
	p.closed = true

	// 关闭所有连接
	close(p.conns)
	for conn := range p.conns {
		conn.Close()
	}

	log.Info("ElectrumX连接池已关闭")
	return nil
}

// Stats 获取连接池统计信息
func (p *ConnPool) Stats() (idleConns, openConns int) {
	p.mu.Lock()
	defer p.mu.Unlock()

	return len(p.conns), p.createdConns
}

// connectionCleaner 定期清理空闲连接
func (p *ConnPool) connectionCleaner() {
	ticker := time.NewTicker(p.idleTimeout / 2)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			p.mu.Lock()
			if p.closed {
				p.mu.Unlock()
				return
			}

			// 获取当前连接数
			connsCount := len(p.conns)
			// 关闭不需要的空闲连接
			if connsCount > p.maxIdleConns {
				toClose := connsCount - p.maxIdleConns
				for i := 0; i < toClose; i++ {
					select {
					case conn := <-p.conns:
						conn.Close()
						p.createdConns--
					default:
						// 没有更多连接可关闭
						break
					}
				}
			}
			p.mu.Unlock()
		case <-p.cleanerCtx.Done():
			return
		}
	}
}
