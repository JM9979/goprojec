package electrumx

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net"
	"sync"
	"sync/atomic"
	"time"

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
	conn      net.Conn
	requestID int32
	mu        sync.Mutex
	config    *RPCConfig
	connected bool
}

// NewClient 创建新的ElectrumX客户端
func NewClient() (*ElectrumXClient, error) {
	// 获取配置
	config := GetRPCConfig()
	if config == nil {
		return nil, fmt.Errorf("获取ElectrumX配置失败")
	}

	client := &ElectrumXClient{
		config: config,
	}

	return client, nil
}

// Connect 连接到ElectrumX服务器
func (c *ElectrumXClient) Connect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.connected && c.conn != nil {
		return nil
	}

	// 构建地址
	address := fmt.Sprintf("%s:%d", c.config.Host, c.config.Port)
	log.Infof("正在连接ElectrumX服务器: %s", address)

	// 设置连接超时
	dialer := &net.Dialer{
		Timeout: c.config.Timeout,
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
		log.Errorf("连接ElectrumX服务器失败: %v", err)
		return fmt.Errorf("连接失败: %w", err)
	}

	c.conn = conn
	c.connected = true
	log.Info("已成功连接到ElectrumX服务器")

	return nil
}

// Disconnect 断开与ElectrumX服务器的连接
func (c *ElectrumXClient) Disconnect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn != nil {
		err := c.conn.Close()
		c.conn = nil
		c.connected = false
		if err != nil {
			log.Errorf("关闭ElectrumX连接时出错: %v", err)
			return fmt.Errorf("关闭连接失败: %w", err)
		}
		log.Info("已断开与ElectrumX服务器的连接")
	}

	return nil
}

// CallRPC 调用ElectrumX RPC方法
func (c *ElectrumXClient) CallRPC(method string, params interface{}) (json.RawMessage, error) {
	// 确保连接
	if err := c.Connect(); err != nil {
		return nil, err
	}

	c.mu.Lock()
	defer c.mu.Unlock()

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
	log.Debugf("发送ElectrumX RPC请求: method=%s, params=%v", method, params)

	// 设置读写超时
	deadline := time.Now().Add(c.config.Timeout)
	if err := c.conn.SetDeadline(deadline); err != nil {
		log.Warnf("设置连接超时失败: %v", err)
	}

	// 发送请求
	_, err = c.conn.Write(reqBytes)
	if err != nil {
		log.Errorf("发送RPC请求失败: %v", err)
		// 尝试重新连接
		c.Disconnect()
		return nil, fmt.Errorf("发送RPC请求失败: %w", err)
	}

	// 读取响应（最大4096字节，应该足够大多数响应）
	buf := make([]byte, 4096)
	n, err := c.conn.Read(buf)
	if err != nil {
		log.Errorf("读取RPC响应失败: %v", err)
		// 连接可能已断开，尝试断开
		c.Disconnect()
		return nil, fmt.Errorf("读取RPC响应失败: %w", err)
	}

	// 解析响应
	var resp RPCResponse
	if err := json.Unmarshal(buf[:n], &resp); err != nil {
		log.Errorf("解析RPC响应失败: %v, 响应内容: %s", err, string(buf[:n]))
		return nil, fmt.Errorf("解析RPC响应失败: %w", err)
	}

	// 检查ID是否匹配
	if resp.ID != id {
		log.Errorf("RPC响应ID不匹配: 期望=%d, 实际=%d", id, resp.ID)
		return nil, fmt.Errorf("RPC响应ID不匹配")
	}

	// 检查错误
	if resp.Error != nil {
		log.Warnf("RPC调用错误: %s (代码: %d)", resp.Error.Message, resp.Error.Code)
		return nil, fmt.Errorf("RPC调用错误: %s (代码: %d)", resp.Error.Message, resp.Error.Code)
	}

	return resp.Result, nil
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

		// 调用同步RPC方法
		result, err := c.CallRPC(method, params)

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
	config := GetRPCConfig()
	if config == nil {
		return fmt.Errorf("获取ElectrumX RPC配置失败")
	}

	if config.Host == "" {
		return fmt.Errorf("ElectrumX RPC Host未配置")
	}

	if config.Port <= 0 {
		return fmt.Errorf("ElectrumX RPC Port无效")
	}

	log.Infof("ElectrumX RPC客户端初始化完成，服务器: %s:%d", config.Host, config.Port)
	return nil
}
