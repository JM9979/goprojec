package blockchain

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"ginproject/entity/config"
	"ginproject/middleware/log"
)

// RPCRequest 表示RPC请求
type RPCRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      string      `json:"id"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params"`
}

// RPCResponse 表示RPC响应
type RPCResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      string      `json:"id"`
	Result  interface{} `json:"result,omitempty"`
	Error   *RPCError   `json:"error,omitempty"`
}

// RPCError 表示RPC错误
type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// AsyncResult 表示异步结果
type AsyncResult struct {
	Result interface{}
	Error  error
}

// 全局连接池
var globalConnPool *ConnPool

// Init 初始化区块链RPC客户端
func Init() error {
	log.Info("初始化区块链RPC客户端...")

	// 获取并验证配置
	config := config.GetConfig().GetTBCNodeConfig()
	if config == nil {
		return fmt.Errorf("获取区块链RPC配置失败")
	}

	if config.URL == "" {
		return fmt.Errorf("区块链RPC URL未配置")
	}

	// 初始化连接池
	pool, err := NewConnPool(nil)
	if err != nil {
		return fmt.Errorf("初始化区块链节点连接池失败: %w", err)
	}
	globalConnPool = pool

	log.Infof("区块链RPC客户端初始化完成，服务器: %s", config.URL)
	return nil
}

// CallRPC 使用上下文同步调用节点RPC
func CallRPC(ctx context.Context, method string, params interface{}, fullResponse bool) (interface{}, error) {
	// 确保连接池已初始化
	if globalConnPool == nil {
		// 获取RPC配置
		config := config.GetConfig().GetTBCNodeConfig()
		if config == nil {
			return nil, fmt.Errorf("RPC配置未初始化")
		}

		// 创建HTTP请求
		req, err := createRPCRequest(config, method, params)
		if err != nil {
			return nil, err
		}

		// 创建带超时的客户端
		client := &http.Client{
			Timeout: time.Duration(config.Timeout) * time.Second,
		}

		// 记录请求日志
		log.Debugf("发送区块链RPC请求: method=%s, params=%v", method, params)

		// 发送请求
		resp, err := client.Do(req.WithContext(ctx))
		if err != nil {
			return nil, fmt.Errorf("发送RPC请求失败: %w", err)
		}
		defer resp.Body.Close()

		return processRPCResponse(resp, fullResponse)
	}

	// 使用连接池处理请求
	result, err := globalConnPool.Call(ctx, method, params)
	if err != nil {
		return nil, err
	}

	if fullResponse {
		return RPCResponse{
			JSONRPC: "1.0",
			ID:      "blockchain_client",
			Result:  result,
		}, nil
	}
	return result, nil
}

// createRPCRequest 创建RPC请求
func createRPCRequest(config *config.TBCNodeConfig, method string, params interface{}) (*http.Request, error) {
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
	req, err := http.NewRequest("POST", config.URL, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("创建HTTP请求失败: %w", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")

	// 设置基本认证
	if config.User != "" && config.Password != "" {
		req.SetBasicAuth(config.User, config.Password)
	}

	return req, nil
}

// processRPCResponse 处理RPC响应
func processRPCResponse(resp *http.Response, fullResponse bool) (interface{}, error) {
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

	// 返回结果
	if fullResponse {
		return rpcResp, nil
	}
	return rpcResp.Result, nil
}

// CallRPCAsync 异步调用节点RPC
func CallRPCAsync(ctx context.Context, method string, params interface{}, fullResponse bool) <-chan AsyncResult {
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

		// 调用同步版本的RPC方法
		result, err := CallRPC(ctx, method, params, fullResponse)

		// 将结果发送到通道
		resultChan <- AsyncResult{
			Result: result,
			Error:  err,
		}
	}()

	return resultChan
}

// Close 关闭客户端和连接池
func Close() error {
	if globalConnPool != nil {
		err := globalConnPool.Close()
		globalConnPool = nil
		return err
	}
	return nil
}
