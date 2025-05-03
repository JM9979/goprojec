package blockchain

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"ginproject/entity/blockchain"
	"ginproject/middleware/log"
)

// DecodeTxHash 根据交易哈希获取交易详情
// 此方法通过区块链节点RPC接口查询指定交易哈希的详细信息
// 返回的交易信息包括输入输出、脚本、金额等详细数据
// 参数:
//   - ctx: 上下文，用于控制请求的生命周期
//   - txid: 交易哈希ID
//
// 返回:
//   - *TransactionResponse: 包含交易详情的结构体指针
//   - error: 如有错误发生则返回错误信息
func DecodeTxHash(ctx context.Context, txid string) (*blockchain.TransactionResponse, error) {
	if txid == "" {
		return nil, errors.New("交易ID不能为空")
	}

	// 记录开始调用日志
	log.Infof("开始查询交易: %s", txid)

	// 调用区块链节点RPC
	result, err := CallRPC("getrawtransaction", []interface{}{txid, 1}, false)
	if err != nil {
		log.Errorf("查询交易失败: %s, 错误: %v", txid, err)
		return nil, fmt.Errorf("解析交易失败: %w", err)
	}

	// 将返回结果转换为TransactionResponse结构
	var tx blockchain.TransactionResponse
	resultBytes, err := json.Marshal(result)
	if err != nil {
		log.Errorf("序列化交易结果失败: %v", err)
		return nil, fmt.Errorf("解析RPC响应失败: %w", err)
	}

	if err := json.Unmarshal(resultBytes, &tx); err != nil {
		log.Errorf("解析交易数据失败: %s, 错误: %v", txid, err)
		return nil, fmt.Errorf("解析交易数据失败: %w", err)
	}

	log.Infof("成功查询交易: %s, 确认数: %d", txid, tx.Confirmations)
	return &tx, nil
}

// DecodeTxHashAsync 异步获取交易详情
// 这是DecodeTxHash的异步版本，用于需要非阻塞调用的场景
// 例如在批量处理多个交易时，可以同时发起多个异步请求以提高性能
func DecodeTxHashAsync(ctx context.Context, txid string) <-chan AsyncResult {
	resultChan := make(chan AsyncResult, 1)

	go func() {
		defer close(resultChan)

		// 调用同步方法
		result, err := DecodeTxHash(ctx, txid)

		// 将结果发送到通道
		resultChan <- AsyncResult{
			Result: result,
			Error:  err,
		}
	}()

	return resultChan
}
