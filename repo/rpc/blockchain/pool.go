package blockchain

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"ginproject/entity/config"
	"ginproject/middleware/log"
)

// HTTPConnection 表示一个HTTP连接
type HTTPConnection struct {
	client    *http.Client
	config    *config.TBCNodeConfig
	lastUsed  time.Time
	isInvalid bool
}

// ConnPool 区块链节点连接池
type ConnPool struct {
	mu            sync.Mutex
	config        *config.TBCNodeConfig
	conns         chan *HTTPConnection
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
func NewConnPool(poolConfig *PoolConfig) (*ConnPool, error) {
	// 获取区块链节点配置
	tbcNodeConfig := config.GetConfig().GetTBCNodeConfig()
	if tbcNodeConfig == nil {
		return nil, fmt.Errorf("获取区块链节点配置失败")
	}

	// 设置默认值
	maxIdleConns := defaultMaxIdleConns
	maxOpenConns := defaultMaxOpenConns
	idleTimeout := defaultIdleTimeout

	// 优先使用配置中的值
	if tbcNodeConfig.MaxIdleConns > 0 {
		maxIdleConns = tbcNodeConfig.MaxIdleConns
	}
	if tbcNodeConfig.MaxOpenConns > 0 {
		maxOpenConns = tbcNodeConfig.MaxOpenConns
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
		config:        tbcNodeConfig,
		conns:         make(chan *HTTPConnection, maxIdleConns),
		maxIdleConns:  maxIdleConns,
		maxOpenConns:  maxOpenConns,
		connTimeout:   time.Duration(tbcNodeConfig.Timeout) * time.Second,
		idleTimeout:   idleTimeout,
		cleanerCtx:    ctx,
		cleanerCancel: cancel,
	}

	// 启动空闲连接清理协程
	go pool.connectionCleaner()

	log.Info("区块链节点连接池已创建, 最大空闲连接:", maxIdleConns,
		", 最大打开连接:", maxOpenConns,
		", 空闲超时:", idleTimeout)
	return pool, nil
}

// GetConn 从池中获取一个连接
func (p *ConnPool) GetConn(ctx context.Context) (*HTTPConnection, error) {
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
		log.Debug("从连接池获取区块链节点连接成功")
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
func (p *ConnPool) createConn(ctx context.Context) (*HTTPConnection, error) {
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

	// 创建HTTP客户端
	client := &http.Client{
		Timeout: p.connTimeout,
	}

	// 创建连接对象
	conn := &HTTPConnection{
		client:    client,
		config:    p.config,
		lastUsed:  time.Now(),
		isInvalid: false,
	}

	// 验证连接是否可用
	if !p.validateConn(conn) {
		p.mu.Lock()
		p.connErr = errors.New("无法创建有效的区块链节点连接")
		p.lastConnErr = time.Now()
		p.createdConns--
		p.mu.Unlock()
		log.ErrorWithContext(ctx, "创建区块链节点连接失败")
		return nil, p.connErr
	}

	log.DebugWithContext(ctx, "创建新的区块链节点连接成功")
	return conn, nil
}

// PutConn 将连接放回池中
func (p *ConnPool) PutConn(conn *HTTPConnection) {
	if conn == nil {
		return
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed || conn.isInvalid {
		p.createdConns--
		return
	}

	// 更新最后使用时间
	conn.lastUsed = time.Now()

	// 尝试将连接放入空闲池
	select {
	case p.conns <- conn:
		// 成功放入空闲池
		return
	default:
		// 空闲池已满，关闭连接
		p.createdConns--
	}
}

// validateConn 验证连接是否有效
func (p *ConnPool) validateConn(conn *HTTPConnection) bool {
	if conn == nil || conn.isInvalid {
		return false
	}

	// 如果连接太久未使用，标记为无效
	if time.Since(conn.lastUsed) > p.idleTimeout {
		conn.isInvalid = true
		return false
	}

	// 构建ping请求
	pingReq := RPCRequest{
		JSONRPC: "1.0",
		ID:      "ping",
		Method:  "ping",
		Params:  []interface{}{},
	}

	pingBytes, err := json.Marshal(pingReq)
	if err != nil {
		log.Error("序列化ping请求失败:", err)
		conn.isInvalid = true
		return false
	}

	// 创建HTTP请求
	req, err := http.NewRequest("POST", conn.config.URL, bytes.NewBuffer(pingBytes))
	if err != nil {
		log.Error("创建HTTP ping请求失败:", err)
		conn.isInvalid = true
		return false
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")

	// 设置基本认证
	if conn.config.User != "" && conn.config.Password != "" {
		req.SetBasicAuth(conn.config.User, conn.config.Password)
	}

	// 设置短超时
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	req = req.WithContext(ctx)

	// 发送请求
	resp, err := conn.client.Do(req)
	if err != nil {
		log.Error("发送ping请求失败:", err)
		conn.isInvalid = true
		return false
	}
	defer resp.Body.Close()

	// 读取并丢弃响应体
	_, err = io.ReadAll(resp.Body)
	if err != nil {
		log.Error("读取ping响应失败:", err)
		conn.isInvalid = true
		return false
	}

	return true
}

// Call 使用连接池中的连接调用RPC方法
func (p *ConnPool) Call(ctx context.Context, method string, params interface{}) (interface{}, error) {
	// 获取连接
	conn, err := p.GetConn(ctx)
	if err != nil {
		return nil, fmt.Errorf("获取连接失败: %w", err)
	}
	defer p.PutConn(conn)

	// 创建RPC请求
	rpcReq := RPCRequest{
		JSONRPC: "1.0",
		ID:      "blockchain_client",
		Method:  method,
		Params:  params,
	}

	// 序列化请求
	reqBody, err := json.Marshal(rpcReq)
	if err != nil {
		return nil, fmt.Errorf("序列化RPC请求失败: %w", err)
	}

	// 创建HTTP请求
	req, err := http.NewRequest("POST", conn.config.URL, bytes.NewBuffer(reqBody))
	if err != nil {
		conn.isInvalid = true
		return nil, fmt.Errorf("创建HTTP请求失败: %w", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")

	// 设置基本认证
	if conn.config.User != "" && conn.config.Password != "" {
		req.SetBasicAuth(conn.config.User, conn.config.Password)
	}

	// 使用上下文
	req = req.WithContext(ctx)

	// 记录请求日志
	log.Debugf("发送区块链RPC请求: method=%s, params=%v", method, params)

	// 发送请求
	resp, err := conn.client.Do(req)
	if err != nil {
		conn.isInvalid = true
		return nil, fmt.Errorf("发送RPC请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取RPC响应失败: %w", err)
	}

	// 解析响应
	var rpcResp RPCResponse
	if err := json.Unmarshal(respBody, &rpcResp); err != nil {
		log.Errorf("解析RPC响应失败: %s", string(respBody))
		return nil, fmt.Errorf("解析RPC响应失败: %w", err)
	}

	// 检查错误
	if rpcResp.Error != nil {
		log.Warnf("RPC调用错误: %s (代码: %d)", rpcResp.Error.Message, rpcResp.Error.Code)
		return nil, fmt.Errorf("RPC调用错误: %s (代码: %d)", rpcResp.Error.Message, rpcResp.Error.Code)
	}

	return rpcResp.Result, nil
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
	for range p.conns {
		p.createdConns--
	}

	log.Info("区块链节点连接池已关闭")
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
					case <-p.conns:
						p.createdConns--
					default:
					}
				}
			}
			p.mu.Unlock()
		case <-p.cleanerCtx.Done():
			return
		}
	}
}
