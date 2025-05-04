package electrumx

import (
	"bufio"
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
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

	// 使用bufio.Reader读取完整响应
	reader := bufio.NewReader(c.conn)

	// 使用bytes.Buffer收集响应数据
	var responseBuffer bytes.Buffer

	// 最大允许的响应大小(10MB)，防止异常大的响应
	const maxResponseSize = 10 * 1024 * 1024

	// 循环读取响应直到获取完整JSON或达到大小限制
	for {
		// 检查是否超出大小限制
		if responseBuffer.Len() > maxResponseSize {
			c.Disconnect()
			log.Errorf("RPC响应超过大小限制(%dMB)", maxResponseSize/1024/1024)
			return nil, fmt.Errorf("RPC响应数据过大，超过%dMB限制", maxResponseSize/1024/1024)
		}

		// 读取一行数据(ElectrumX响应通常以换行符结束)
		line, err := reader.ReadBytes('\n')
		if err != nil && err != io.EOF {
			log.Errorf("读取RPC响应失败: %v", err)
			c.Disconnect()
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
				// 检查错误
				if resp.Error != nil {
					log.Warnf("RPC调用错误: %s (代码: %d)", resp.Error.Message, resp.Error.Code)
					return nil, fmt.Errorf("RPC调用错误: %s (代码: %d)", resp.Error.Message, resp.Error.Code)
				}

				log.Debugf("成功接收ElectrumX RPC响应: method=%s, 大小=%d字节", method, responseBuffer.Len())
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

			log.Errorf("连接关闭但未收到完整响应: %s", respData)
			c.Disconnect()
			return nil, fmt.Errorf("连接关闭但未收到完整响应")
		}
	}
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
