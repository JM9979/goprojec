package rpc

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"ginproject/entity/config"
	"ginproject/middleware/conf"
	"ginproject/middleware/log"
)

// DefaultTimeout 是RPC调用的默认超时时间
const DefaultTimeout = 30 * time.Second

var (
	// Config 全局RPC配置
	Config config.RPCConfig
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

// Init 初始化RPC客户端
func Init() error {
	// 从配置中获取RPC相关设置
	rpcConfig := conf.GetRPCConfig()

	// 验证必要的配置项
	if rpcConfig.URL == "" {
		return fmt.Errorf("缺少RPC URL配置")
	}

	if rpcConfig.User == "" || rpcConfig.Password == "" {
		log.Warn("RPC认证信息不完整，可能需要在某些环境中进行认证")
	}

	// 初始化配置
	Config = *rpcConfig

	log.Info("RPC客户端初始化成功")
	return nil
}

// CallNodeRPC 同步调用节点RPC
func CallNodeRPC(method string, params interface{}, fullResponse bool) (interface{}, error) {
	// 创建RPC请求
	rpcReq := RPCRequest{
		JSONRPC: "1.0",
		ID:      conf.GetServerConfig().Name,
		Method:  method,
		Params:  params,
	}

	// 序列化请求
	reqBody, err := json.Marshal(rpcReq)
	if err != nil {
		return nil, fmt.Errorf("序列化RPC请求失败: %w", err)
	}

	// 创建HTTP请求
	req, err := http.NewRequest("POST", Config.URL, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("创建HTTP请求失败: %w", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	// 设置基本认证
	req.SetBasicAuth(Config.User, Config.Password)

	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("发送RPC请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取RPC响应失败: %w", err)
	}

	// 解析响应
	var rpcResp RPCResponse
	if err := json.Unmarshal(respBody, &rpcResp); err != nil {
		return nil, fmt.Errorf("解析RPC响应失败: %w", err)
	}

	// 检查错误
	if rpcResp.Error != nil {
		return nil, fmt.Errorf("RPC调用错误: %s (代码: %d)", rpcResp.Error.Message, rpcResp.Error.Code)
	}

	// 返回结果
	if fullResponse {
		return rpcResp, nil
	}
	return rpcResp.Result, nil
}

// CallNodeRPCAsync 异步调用节点RPC
func CallNodeRPCAsync(ctx context.Context, method string, params interface{}, fullResponse bool) <-chan AsyncResult {
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
		result, err := CallNodeRPC(method, params, fullResponse)

		// 将结果发送到通道
		resultChan <- AsyncResult{
			Result: result,
			Error:  err,
		}
	}()

	return resultChan
}
