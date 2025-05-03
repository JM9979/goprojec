package electrumx

import (
	"context"
	"encoding/json"
	"fmt"

	"ginproject/entity/electrumx"
	"ginproject/middleware/log"
)

// Global client instance
var (
	defaultClient *ElectrumXClient
	initialized   = false
)

// GetDefaultClient 获取默认的ElectrumX客户端实例
func GetDefaultClient() (*ElectrumXClient, error) {
	if !initialized {
		if err := Init(); err != nil {
			return nil, err
		}

		client, err := NewClient()
		if err != nil {
			return nil, err
		}
		defaultClient = client
		initialized = true
	}

	return defaultClient, nil
}

// CallMethod 调用ElectrumX RPC方法的简便函数
func CallMethod(method string, params []interface{}) (json.RawMessage, error) {
	client, err := GetDefaultClient()
	if err != nil {
		return nil, fmt.Errorf("获取ElectrumX客户端失败: %w", err)
	}

	// 记录开始调用日志
	log.Infof("开始调用ElectrumX方法: %s", method)

	result, err := client.CallRPC(method, params)
	if err != nil {
		log.Errorf("调用ElectrumX方法失败: %s, 错误: %v", method, err)
		return nil, err
	}

	return result, nil
}

// CallMethodAsync 异步调用ElectrumX RPC方法的简便函数
func CallMethodAsync(ctx context.Context, method string, params []interface{}) <-chan AsyncResult {
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

		result, err := CallMethod(method, params)
		resultChan <- AsyncResult{
			Result: result,
			Error:  err,
		}
	}()

	return resultChan
}

// 以下是常用的ElectrumX RPC方法

// GetBlockHeader 获取区块头信息
func GetBlockHeader(height int) (map[string]interface{}, error) {
	result, err := CallMethod("blockchain.block.header", []interface{}{height})
	if err != nil {
		return nil, err
	}

	var header map[string]interface{}
	if err := json.Unmarshal(result, &header); err != nil {
		log.Errorf("解析区块头失败: %v", err)
		return nil, fmt.Errorf("解析区块头失败: %w", err)
	}

	return header, nil
}

// GetBalance 获取地址余额
func GetBalance(address string) (map[string]interface{}, error) {
	result, err := CallMethod("blockchain.scripthash.get_balance", []interface{}{address})
	if err != nil {
		return nil, err
	}

	var balance map[string]interface{}
	if err := json.Unmarshal(result, &balance); err != nil {
		log.Errorf("解析余额失败: %v", err)
		return nil, fmt.Errorf("解析余额失败: %w", err)
	}

	return balance, nil
}

// GetTransactionHistory 获取地址交易历史
func GetTransactionHistory(address string) ([]interface{}, error) {
	result, err := CallMethod("blockchain.scripthash.get_history", []interface{}{address})
	if err != nil {
		return nil, err
	}

	var history []interface{}
	if err := json.Unmarshal(result, &history); err != nil {
		log.Errorf("解析交易历史失败: %v", err)
		return nil, fmt.Errorf("解析交易历史失败: %w", err)
	}

	return history, nil
}

// BroadcastTransaction 广播原始交易
func BroadcastTransaction(rawTx string) (string, error) {
	result, err := CallMethod("blockchain.transaction.broadcast", []interface{}{rawTx})
	if err != nil {
		return "", err
	}

	var txid string
	if err := json.Unmarshal(result, &txid); err != nil {
		log.Errorf("解析交易ID失败: %v", err)
		return "", fmt.Errorf("解析交易ID失败: %w", err)
	}

	return txid, nil
}

// GetTransaction 获取交易详情
func GetTransaction(txid string) (map[string]interface{}, error) {
	result, err := CallMethod("blockchain.transaction.get", []interface{}{txid, true})
	if err != nil {
		return nil, err
	}

	var tx map[string]interface{}
	if err := json.Unmarshal(result, &tx); err != nil {
		log.Errorf("解析交易详情失败: %v", err)
		return nil, fmt.Errorf("解析交易详情失败: %w", err)
	}

	return tx, nil
}

// EstimateFee 估算交易手续费
func EstimateFee(blocks int) (float64, error) {
	result, err := CallMethod("blockchain.estimatefee", []interface{}{blocks})
	if err != nil {
		return 0, err
	}

	var fee float64
	if err := json.Unmarshal(result, &fee); err != nil {
		log.Errorf("解析手续费失败: %v", err)
		return 0, fmt.Errorf("解析手续费失败: %w", err)
	}

	return fee, nil
}

// ServerVersion 获取服务器版本信息
func ServerVersion() ([]interface{}, error) {
	result, err := CallMethod("server.version", []interface{}{"electrumx-client", "1.4"})
	if err != nil {
		return nil, err
	}

	var version []interface{}
	if err := json.Unmarshal(result, &version); err != nil {
		log.Errorf("解析服务器版本信息失败: %v", err)
		return nil, fmt.Errorf("解析服务器版本信息失败: %w", err)
	}

	return version, nil
}

// ServerPeers 获取服务器的对等节点信息
func ServerPeers() ([]interface{}, error) {
	result, err := CallMethod("server.peers.subscribe", []interface{}{})
	if err != nil {
		return nil, err
	}

	var peers []interface{}
	if err := json.Unmarshal(result, &peers); err != nil {
		log.Errorf("解析对等节点信息失败: %v", err)
		return nil, fmt.Errorf("解析对等节点信息失败: %w", err)
	}

	return peers, nil
}

// ServerFeatures 获取服务器特性
func ServerFeatures() (map[string]interface{}, error) {
	result, err := CallMethod("server.features", []interface{}{})
	if err != nil {
		return nil, err
	}

	var features map[string]interface{}
	if err := json.Unmarshal(result, &features); err != nil {
		log.Errorf("解析服务器特性失败: %v", err)
		return nil, fmt.Errorf("解析服务器特性失败: %w", err)
	}

	return features, nil
}

// GetScriptHashHistory 获取指定脚本哈希的交易历史
func GetScriptHashHistory(scriptHash string) (electrumx.ElectrumXHistoryResponse, error) {
	// 参数校验
	if len(scriptHash) == 0 {
		log.Errorf("调用GetScriptHashHistory失败: scriptHash不能为空")
		return nil, fmt.Errorf("scriptHash不能为空")
	}

	// 记录开始调用日志
	log.Infof("开始获取脚本哈希历史: %s", scriptHash)

	// 调用RPC方法
	result, err := CallMethod("blockchain.scripthash.get_history", []interface{}{scriptHash})
	if err != nil {
		log.Errorf("获取脚本哈希历史失败: %v", err)
		return nil, fmt.Errorf("获取脚本哈希历史失败: %w", err)
	}

	// 解析响应
	var history electrumx.ElectrumXHistoryResponse
	if err := json.Unmarshal(result, &history); err != nil {
		log.Errorf("解析脚本哈希历史失败: %v, 原始数据: %s", err, string(result))
		return nil, fmt.Errorf("解析脚本哈希历史失败: %w", err)
	}

	log.Infof("成功获取脚本哈希历史, 共 %d 条记录", len(history))
	return history, nil
}

// GetScriptHashHistoryAsync 异步获取指定脚本哈希的交易历史
func GetScriptHashHistoryAsync(ctx context.Context, scriptHash string) <-chan AsyncHistoryResult {
	resultChan := make(chan AsyncHistoryResult, 1)

	go func() {
		defer close(resultChan)

		// 检查上下文是否已取消
		select {
		case <-ctx.Done():
			resultChan <- AsyncHistoryResult{
				Result: nil,
				Error:  ctx.Err(),
			}
			return
		default:
			// 继续执行
		}

		history, err := GetScriptHashHistory(scriptHash)
		resultChan <- AsyncHistoryResult{
			Result: history,
			Error:  err,
		}
	}()

	return resultChan
}

// AsyncHistoryResult 表示异步获取历史记录的结果
type AsyncHistoryResult struct {
	Result electrumx.ElectrumXHistoryResponse
	Error  error
}

// GetScriptHistory 获取脚本历史
func GetScriptHistory(ctx context.Context, scriptHash string) (electrumx.ElectrumXHistoryResponse, error) {
	// 参数校验
	if len(scriptHash) == 0 {
		log.ErrorWithContextf(ctx, "GetScriptHistory失败: scriptHash不能为空")
		return nil, fmt.Errorf("scriptHash不能为空")
	}

	// 记录开始调用日志
	log.InfoWithContextf(ctx, "开始获取脚本哈希历史: %s", scriptHash)

	// 调用RPC方法
	result, err := CallMethod("blockchain.scripthash.get_history", []interface{}{scriptHash})
	if err != nil {
		log.ErrorWithContextf(ctx, "获取脚本哈希历史失败: %v", err)
		return nil, fmt.Errorf("获取脚本哈希历史失败: %w", err)
	}

	// 解析响应
	var history electrumx.ElectrumXHistoryResponse
	if err := json.Unmarshal(result, &history); err != nil {
		log.ErrorWithContextf(ctx, "解析脚本哈希历史失败: %v, 原始数据: %s", err, string(result))
		return nil, fmt.Errorf("解析脚本哈希历史失败: %w", err)
	}

	log.InfoWithContextf(ctx, "成功获取脚本哈希历史, 共 %d 条记录", len(history))
	return history, nil
}
