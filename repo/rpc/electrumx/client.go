package electrumx

import (
	"bufio"
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"ginproject/entity/config"
	"ginproject/middleware/log"
)

// RPCRequest 表示RPC请求
type RPCRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      int         `json:"id"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params"`
}

// RPCResponse 表示RPC响应
type RPCResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      int             `json:"id"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *RPCError       `json:"error,omitempty"`
}

// RPCError 表示RPC错误
type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// AsyncResult 表示异步结果
type AsyncResult struct {
	Result json.RawMessage
	Error  error
}

// ElectrumXClient ElectrumX客户端
type ElectrumXClient struct {
	// mu        sync.Mutex
	requestID int32
	config    *config.ElectrumXConfig

	// 连接池相关字段
	pool    *ConnPool
	poolMu  sync.Mutex
	usePool bool
}

// 错误定义
var (
	// ErrNoPool 未启用连接池错误
	ErrNoPool = errors.New("连接池未启用")
)

// NewClient 创建新的ElectrumX客户端
func NewClient() (*ElectrumXClient, error) {
	// 获取配置
	config := config.GetConfig().GetElectrumXConfig()
	if config == nil {
		return nil, fmt.Errorf("获取ElectrumX配置失败")
	}

	client := &ElectrumXClient{
		config:  config,
		usePool: true, // 默认启用连接池
	}

	// 自动初始化连接池
	err := client.EnablePool()
	if err != nil {
		log.Warn("自动启用ElectrumX连接池失败:", err, "将使用直接连接模式")
		client.usePool = false
	}

	return client, nil
}

// EnablePool 启用连接池
func (c *ElectrumXClient) EnablePool() error {
	c.poolMu.Lock()
	defer c.poolMu.Unlock()

	if c.pool != nil {
		// 连接池已存在，直接启用
		c.usePool = true
		return nil
	}

	// 创建连接池
	pool, err := NewConnPool(c, &PoolConfig{
		MaxIdleConns: c.config.MaxIdleConns,
		MaxOpenConns: c.config.MaxOpenConns,
		IdleTimeout:  0, // 使用默认值
	})
	if err != nil {
		return fmt.Errorf("创建连接池失败: %w", err)
	}

	c.pool = pool
	c.usePool = true
	log.Info("ElectrumX客户端已启用连接池")
	return nil
}

// DisablePool 禁用连接池
func (c *ElectrumXClient) DisablePool() error {
	c.poolMu.Lock()
	defer c.poolMu.Unlock()

	c.usePool = false

	if c.pool != nil {
		if err := c.pool.Close(); err != nil {
			return fmt.Errorf("关闭连接池失败: %w", err)
		}
		c.pool = nil
	}

	log.Info("ElectrumX客户端已禁用连接池")
	return nil
}

// PoolStats 获取连接池统计信息
func (c *ElectrumXClient) PoolStats() (idleConns, openConns int, err error) {
	c.poolMu.Lock()
	defer c.poolMu.Unlock()

	if c.pool == nil || !c.usePool {
		return 0, 0, ErrNoPool
	}

	idleConns, openConns = c.pool.Stats()
	return idleConns, openConns, nil
}

// Connect 连接到ElectrumX服务器，返回一个新连接由调用者管理
func (c *ElectrumXClient) Connect() (net.Conn, error) {
	// 创建一个新连接
	address := net.JoinHostPort(c.config.Host, fmt.Sprintf("%d", c.config.Port))
	log.Info("正在创建ElectrumX连接, 服务器地址:", address)

	// 设置连接超时
	dialer := &net.Dialer{
		Timeout: time.Duration(c.config.Timeout) * time.Second,
	}

	var conn net.Conn
	var err error

	// 根据是否使用TLS创建连接
	if c.config.UseTLS {
		tlsConfig := &tls.Config{
			InsecureSkipVerify: true, // 注意：生产环境中应该验证证书
		}
		conn, err = tls.DialWithDialer(dialer, c.config.Protocol, address, tlsConfig)
	} else {
		conn, err = dialer.Dial(c.config.Protocol, address)
	}

	if err != nil {
		log.Error("创建ElectrumX连接失败:", err)
		return nil, fmt.Errorf("创建连接失败: %w", err)
	}

	// 测试连接
	deadline := time.Now().Add(time.Duration(c.config.Timeout) * time.Second)
	if err := conn.SetDeadline(deadline); err != nil {
		conn.Close()
		return nil, fmt.Errorf("设置连接超时失败: %w", err)
	}

	// 构建ping请求
	id := int(atomic.AddInt32(&c.requestID, 1))
	pingReq := RPCRequest{
		JSONRPC: "2.0",
		ID:      id,
		Method:  "server.ping",
		Params:  []interface{}{},
	}

	pingBytes, err := json.Marshal(pingReq)
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("序列化ping请求失败: %w", err)
	}
	pingBytes = append(pingBytes, '\n')

	// 发送ping请求
	if _, err = conn.Write(pingBytes); err != nil {
		conn.Close()
		return nil, fmt.Errorf("发送ping请求失败: %w", err)
	}

	// 读取响应
	reader := bufio.NewReader(conn)
	_, err = reader.ReadBytes('\n')
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("读取ping响应失败: %w", err)
	}

	// 重置超时
	conn.SetDeadline(time.Time{})

	log.Info("成功创建并测试了ElectrumX连接")
	return conn, nil
}

// CallRPC 调用ElectrumX RPC方法
func (c *ElectrumXClient) CallRPC(method string, params interface{}) (json.RawMessage, error) {
	// 检查是否使用连接池
	c.poolMu.Lock()
	usePool := c.usePool && c.pool != nil
	c.poolMu.Unlock()

	if usePool {
		// 使用连接池调用
		return c.callRPCWithPool(context.Background(), method, params)
	}

	// 使用传统方式调用
	return c.callRPCDirect(method, params)
}

// callRPCDirect 使用直接连接方式调用RPC
func (c *ElectrumXClient) callRPCDirect(method string, params interface{}) (json.RawMessage, error) {
	// 创建新连接
	conn, err := c.Connect()
	if err != nil {
		return nil, err
	}
	// 确保连接最终会被关闭
	defer conn.Close()

	// 获取请求ID
	id := int(atomic.AddInt32(&c.requestID, 1))

	// 构建请求
	req := RPCRequest{
		JSONRPC: "2.0",
		ID:      id,
		Method:  method,
		Params:  params,
	}

	// 编码请求
	reqBytes, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("序列化RPC请求失败: %w", err)
	}
	reqBytes = append(reqBytes, '\n') // ElectrumX要求每个请求以换行符结束

	// 记录日志
	log.Debug("发送ElectrumX RPC请求:", "method:", method, "params:", params)

	// 设置读写超时
	deadline := time.Now().Add(time.Duration(c.config.Timeout) * time.Second)
	if err := conn.SetDeadline(deadline); err != nil {
		log.Warn("设置连接超时失败:", err)
		return nil, fmt.Errorf("设置连接超时失败: %w", err)
	}

	// 发送请求
	_, err = conn.Write(reqBytes)
	if err != nil {
		log.Error("发送RPC请求失败:", err)
		return nil, fmt.Errorf("发送RPC请求失败: %w", err)
	}

	// 使用bufio.Reader读取完整响应
	reader := bufio.NewReader(conn)

	// 使用bytes.Buffer收集响应数据
	var responseBuffer bytes.Buffer

	// 最大允许的响应大小(10MB)，防止异常大的响应
	const maxResponseSize = 10 * 1024 * 1024

	// 循环读取响应直到获取完整JSON或达到大小限制
	for {
		// 检查是否超出大小限制
		if responseBuffer.Len() > maxResponseSize {
			log.Error("RPC响应超过大小限制(", maxResponseSize/1024/1024, "MB)")
			return nil, fmt.Errorf("RPC响应数据过大，超过%dMB限制", maxResponseSize/1024/1024)
		}

		// 读取一行数据(ElectrumX响应通常以换行符结束)
		line, err := reader.ReadBytes('\n')
		if err != nil && err != io.EOF {
			log.Error("读取RPC响应失败:", err)
			return nil, fmt.Errorf("读取RPC响应失败: %w", err)
		}

		// 将读取的数据添加到缓冲区
		if len(line) > 0 {
			responseBuffer.Write(line)
		}

		// 尝试解析已收集的数据
		responseData := responseBuffer.Bytes()
		var resp RPCResponse

		if jsonErr := json.Unmarshal(responseData, &resp); jsonErr == nil {
			// 检查ID是否匹配
			if resp.ID == id {
				// 重置超时
				conn.SetDeadline(time.Time{})

				// 检查错误
				if resp.Error != nil {
					log.Warn("RPC调用错误:", resp.Error.Message, "(代码:", resp.Error.Code, ")")
					return nil, fmt.Errorf("RPC调用错误: %s (代码: %d)", resp.Error.Message, resp.Error.Code)
				}

				log.Debug("成功接收ElectrumX RPC响应:", "method:", method, "大小:", responseBuffer.Len(), "字节")
				return resp.Result, nil
			}
		}

		// 如果遇到EOF且尚未解析成功，说明连接已关闭但数据可能不完整
		if err == io.EOF {
			// 记录截断的响应数据(最多500字节)
			respData := responseBuffer.String()
			if len(respData) > 500 {
				respData = respData[:500] + "... [截断]"
			}

			log.Error("连接关闭但未收到完整响应:", respData)
			return nil, fmt.Errorf("连接关闭但未收到完整响应")
		}
	}
}

// callRPCWithPool 使用连接池调用RPC
func (c *ElectrumXClient) callRPCWithPool(ctx context.Context, method string, params interface{}) (json.RawMessage, error) {
	c.poolMu.Lock()
	pool := c.pool
	c.poolMu.Unlock()

	if pool == nil {
		return nil, ErrNoPool
	}

	// 获取一个连接
	conn, err := pool.GetConn(ctx)
	if err != nil {
		log.Error("从连接池获取连接失败:", err)
		return nil, fmt.Errorf("从连接池获取连接失败: %w", err)
	}

	// 确保连接最终会归还到池中
	defer pool.PutConn(conn)

	// 获取请求ID
	id := int(atomic.AddInt32(&c.requestID, 1))

	// 构建请求
	req := RPCRequest{
		JSONRPC: "2.0",
		ID:      id,
		Method:  method,
		Params:  params,
	}

	// 编码请求
	reqBytes, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("序列化RPC请求失败: %w", err)
	}
	reqBytes = append(reqBytes, '\n') // ElectrumX要求每个请求以换行符结束

	// 记录日志
	log.Debug("通过连接池发送ElectrumX RPC请求:", "method:", method, "params:", params)

	// 设置读写超时
	deadline := time.Now().Add(time.Duration(c.config.Timeout) * time.Second)
	if err := conn.SetDeadline(deadline); err != nil {
		log.Warn("设置连接超时失败:", err)
		return nil, fmt.Errorf("设置连接超时失败: %w", err)
	}

	// 发送请求
	_, err = conn.Write(reqBytes)
	if err != nil {
		log.Error("发送RPC请求失败:", err)
		return nil, fmt.Errorf("发送RPC请求失败: %w", err)
	}

	// 使用bufio.Reader读取完整响应
	reader := bufio.NewReader(conn)

	// 使用bytes.Buffer收集响应数据
	var responseBuffer bytes.Buffer

	// 最大允许的响应大小(10MB)，防止异常大的响应
	const maxResponseSize = 10 * 1024 * 1024

	// 循环读取响应直到获取完整JSON或达到大小限制
	for {
		// 检查是否超出大小限制
		if responseBuffer.Len() > maxResponseSize {
			log.Error("RPC响应超过大小限制(", maxResponseSize/1024/1024, "MB)")
			return nil, fmt.Errorf("RPC响应数据过大，超过%dMB限制", maxResponseSize/1024/1024)
		}

		// 读取一行数据(ElectrumX响应通常以换行符结束)
		line, err := reader.ReadBytes('\n')
		if err != nil && err != io.EOF {
			log.Error("读取RPC响应失败:", err)
			return nil, fmt.Errorf("读取RPC响应失败: %w", err)
		}

		// 将读取的数据添加到缓冲区
		if len(line) > 0 {
			responseBuffer.Write(line)
		}

		// 尝试解析已收集的数据
		responseData := responseBuffer.Bytes()
		var resp RPCResponse

		if jsonErr := json.Unmarshal(responseData, &resp); jsonErr == nil {
			// 检查ID是否匹配
			if resp.ID == id {
				// 重置超时
				conn.SetDeadline(time.Time{})

				// 检查错误
				if resp.Error != nil {
					log.Warn("RPC调用错误:", resp.Error.Message, "(代码:", resp.Error.Code, ")")
					return nil, fmt.Errorf("RPC调用错误: %s (代码: %d)", resp.Error.Message, resp.Error.Code)
				}

				log.Debug("成功接收ElectrumX RPC响应:", "method:", method, "大小:", responseBuffer.Len(), "字节")
				return resp.Result, nil
			}
		}

		// 如果遇到EOF且尚未解析成功，说明连接已关闭但数据可能不完整
		if err == io.EOF {
			// 记录截断的响应数据(最多500字节)
			respData := responseBuffer.String()
			if len(respData) > 500 {
				respData = respData[:500] + "... [截断]"
			}

			log.Error("连接关闭但未收到完整响应:", respData)
			return nil, fmt.Errorf("连接关闭但未收到完整响应")
		}
	}
}

// CallRPCWithContext 带上下文的RPC调用，支持连接池
func (c *ElectrumXClient) CallRPCWithContext(ctx context.Context, method string, params interface{}) (json.RawMessage, error) {
	// 检查是否使用连接池
	c.poolMu.Lock()
	usePool := c.usePool && c.pool != nil
	c.poolMu.Unlock()

	if usePool {
		// 使用连接池调用
		return c.callRPCWithPool(ctx, method, params)
	}

	// 对于非连接池调用，忽略上下文，使用传统方式
	// 在实际应用中可以进一步完善支持上下文的非连接池调用
	return c.callRPCDirect(method, params)
}

// CallRPCAsync 异步调用ElectrumX RPC方法
func (c *ElectrumXClient) CallRPCAsync(ctx context.Context, method string, params interface{}) <-chan AsyncResult {
	resultChan := make(chan AsyncResult, 1)

	go func() {
		defer close(resultChan)

		// 检查上下文是否已取消
		select {
		case <-ctx.Done():
			resultChan <- AsyncResult{
				Result: nil,
				Error:  ctx.Err(),
			}
			return
		default:
			// 继续执行
		}

		// 调用带上下文的RPC方法
		result, err := c.CallRPCWithContext(ctx, method, params)

		// 将结果发送到通道
		resultChan <- AsyncResult{
			Result: result,
			Error:  err,
		}
	}()

	return resultChan
}

// Init 初始化ElectrumX RPC客户端
func Init() error {
	log.Info("初始化ElectrumX RPC客户端...")

	// 获取并验证配置
	config := config.GetConfig().GetElectrumXConfig()
	if config == nil {
		return fmt.Errorf("获取ElectrumX RPC配置失败")
	}

	if config.Host == "" {
		return fmt.Errorf("ElectrumX RPC Host未配置")
	}

	if config.Port <= 0 {
		return fmt.Errorf("ElectrumX RPC Port无效")
	}

	// 设置默认的连接池大小（如果配置中未指定）
	if config.MaxIdleConns <= 0 {
		config.MaxIdleConns = defaultMaxIdleConns
		log.Info("ElectrumX连接池最大空闲连接数未配置，使用默认值:", defaultMaxIdleConns)
	}

	if config.MaxOpenConns <= 0 {
		config.MaxOpenConns = defaultMaxOpenConns
		log.Info("ElectrumX连接池最大连接数未配置，使用默认值:", defaultMaxOpenConns)
	}

	log.Info("ElectrumX RPC客户端初始化完成，服务器:", net.JoinHostPort(config.Host, fmt.Sprintf("%d", config.Port)),
		"，连接池最大空闲连接:", config.MaxIdleConns, "，最大连接数:", config.MaxOpenConns)
	return nil
}
