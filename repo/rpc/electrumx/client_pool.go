package electrumx

import (
	"bufio"
	"bytes"
	"context"
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

// ElectrumXClientPool ElectrumX客户端池
type ElectrumXClientPool struct {
	mu                sync.Mutex
	config            *config.ElectrumXConfig
	pool              *ConnPool
	client            *ElectrumXClient
	defaultCtxTimeout int
}

var (
	// 单例客户端池
	clientPool     *ElectrumXClientPool
	clientPoolOnce sync.Once
)

var (
	// ErrNoClientAvailable 无可用客户端错误
	ErrNoClientAvailable = errors.New("没有可用的ElectrumX客户端")
)

// GetClientPool 获取ElectrumX客户端池的单例实例
func GetClientPool() (*ElectrumXClientPool, error) {
	var initErr error

	clientPoolOnce.Do(func() {
		log.Info("初始化ElectrumX客户端池...")

		client, err := NewClient()
		if err != nil {
			initErr = fmt.Errorf("创建ElectrumX客户端失败: %w", err)
			return
		}

		// 获取配置
		cfg := config.GetConfig().GetElectrumXConfig()
		if cfg == nil {
			initErr = fmt.Errorf("获取ElectrumX配置失败")
			return
		}

		// 创建连接池
		pool, err := NewConnPool(client, &PoolConfig{
			MaxIdleConns: cfg.MaxIdleConns,
			MaxOpenConns: cfg.MaxOpenConns,
			IdleTimeout:  0, // 使用默认值
		})
		if err != nil {
			initErr = fmt.Errorf("创建ElectrumX连接池失败: %w", err)
			return
		}

		// 创建客户端池
		clientPool = &ElectrumXClientPool{
			config:            cfg,
			pool:              pool,
			client:            client,
			defaultCtxTimeout: cfg.Timeout,
		}

		log.Info("ElectrumX客户端池初始化完成")
	})

	if initErr != nil {
		return nil, initErr
	}

	return clientPool, nil
}

// CallWithConnection 使用连接池中的连接执行方法
func (cp *ElectrumXClientPool) CallWithConnection(ctx context.Context,
	fn func(ctx context.Context, conn net.Conn) (interface{}, error)) (interface{}, error) {

	if cp.pool == nil {
		return nil, ErrNoClientAvailable
	}

	// 获取一个连接
	conn, err := cp.pool.GetConn(ctx)
	if err != nil {
		log.ErrorWithContext(ctx, "从连接池获取连接失败:", err)
		return nil, fmt.Errorf("从连接池获取连接失败: %w", err)
	}

	// 确保连接最终会归还到池中
	defer cp.pool.PutConn(conn)

	// 执行回调函数
	log.DebugWithContext(ctx, "使用连接池连接执行ElectrumX调用")
	result, err := fn(ctx, conn)
	if err != nil {
		log.ErrorWithContext(ctx, "执行ElectrumX调用失败:", err)
		return nil, err
	}

	return result, nil
}

// Close 关闭客户端池及其资源
func (cp *ElectrumXClientPool) Close() error {
	cp.mu.Lock()
	defer cp.mu.Unlock()

	if cp.pool != nil {
		log.Info("关闭ElectrumX客户端池")
		return cp.pool.Close()
	}
	return nil
}

// Stats 获取客户端池状态
func (cp *ElectrumXClientPool) Stats() (idleConns, openConns int) {
	cp.mu.Lock()
	defer cp.mu.Unlock()

	if cp.pool != nil {
		return cp.pool.Stats()
	}
	return 0, 0
}

// RawCallRPC 使用连接池中的连接执行RPC调用
func (cp *ElectrumXClientPool) RawCallRPC(ctx context.Context,
	method string, params interface{}) (json.RawMessage, error) {

	result, err := cp.CallWithConnection(ctx, func(ctx context.Context, conn net.Conn) (interface{}, error) {
		// 获取请求ID
		id := int(atomic.AddInt32(&cp.client.requestID, 1))

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
		log.DebugWithContext(ctx, "发送ElectrumX RPC请求:", "method:", method, "params:", params)

		// 设置读写超时
		deadline := time.Now().Add(time.Duration(cp.defaultCtxTimeout) * time.Second)
		if err := conn.SetDeadline(deadline); err != nil {
			log.WarnWithContext(ctx, "设置连接超时失败:", err)
			return nil, fmt.Errorf("设置连接超时失败: %w", err)
		}

		// 发送请求
		_, err = conn.Write(reqBytes)
		if err != nil {
			log.ErrorWithContext(ctx, "发送RPC请求失败:", err)
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
				log.ErrorWithContext(ctx, "RPC响应超过大小限制(", maxResponseSize/1024/1024, "MB)")
				return nil, fmt.Errorf("RPC响应数据过大，超过%dMB限制", maxResponseSize/1024/1024)
			}

			// 读取一行数据(ElectrumX响应通常以换行符结束)
			line, err := reader.ReadBytes('\n')
			if err != nil && err != io.EOF {
				log.ErrorWithContext(ctx, "读取RPC响应失败:", err)
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
						log.WarnWithContext(ctx, "RPC调用错误:", resp.Error.Message, "(代码:", resp.Error.Code, ")")
						return nil, fmt.Errorf("RPC调用错误: %s (代码: %d)", resp.Error.Message, resp.Error.Code)
					}

					log.DebugWithContext(ctx, "成功接收ElectrumX RPC响应:",
						"method:", method, "大小:", responseBuffer.Len(), "字节")
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

				log.ErrorWithContext(ctx, "连接关闭但未收到完整响应:", respData)
				return nil, fmt.Errorf("连接关闭但未收到完整响应")
			}
		}
	})

	if err != nil {
		return nil, err
	}

	return result.(json.RawMessage), nil
}

// CallPoolRPC 使用连接池调用RPC方法
func CallPoolRPC(ctx context.Context, method string, params interface{}) (json.RawMessage, error) {
	pool, err := GetClientPool()
	if err != nil {
		return nil, fmt.Errorf("获取ElectrumX客户端池失败: %w", err)
	}

	// 记录开始调用日志
	log.InfoWithContext(ctx, "开始调用ElectrumX方法(连接池):", method)

	result, err := pool.RawCallRPC(ctx, method, params)
	if err != nil {
		log.ErrorWithContext(ctx, "调用ElectrumX方法失败:", method, "错误:", err)
		return nil, err
	}

	return result, nil
}

// ResetClientPool 重置客户端池(用于测试)
func ResetClientPool() {
	if clientPool != nil {
		clientPool.Close()
		clientPool = nil
	}
	clientPoolOnce = sync.Once{}
}
