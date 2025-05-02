package electrumx

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"ginproject/middleware/log"
)

// 以下是使用ElectrumX RPC客户端的示例函数

// GetAddressBalance 获取地址余额示例
func GetAddressBalance(address string) (float64, error) {
	// 记录开始日志
	log.Infof("开始查询地址余额: %s", address)

	// 获取地址的余额
	balance, err := GetBalance(address)
	if err != nil {
		log.Errorf("获取地址余额失败: %s, 错误: %v", address, err)
		return 0, err
	}

	// 解析confirmed余额
	confirmed, ok := balance["confirmed"].(float64)
	if !ok {
		log.Warnf("余额格式不正确: %v", balance)
		return 0, fmt.Errorf("余额格式不正确")
	}

	log.Infof("地址 %s 的确认余额为: %f", address, confirmed)
	return confirmed, nil
}

// GetAddressTransactions 获取地址交易历史示例
func GetAddressTransactions(address string, limit int) ([]map[string]interface{}, error) {
	// 记录开始日志
	log.Infof("开始查询地址交易历史: %s, 限制: %d条", address, limit)

	// 获取地址的交易历史
	history, err := GetTransactionHistory(address)
	if err != nil {
		log.Errorf("获取地址交易历史失败: %s, 错误: %v", address, err)
		return nil, err
	}

	// 限制结果数量
	if len(history) > limit && limit > 0 {
		history = history[:limit]
	}

	// 构建结果
	result := make([]map[string]interface{}, 0, len(history))

	// 转换每个交易
	for _, tx := range history {
		txMap, ok := tx.(map[string]interface{})
		if !ok {
			log.Warnf("交易格式不正确: %v", tx)
			continue
		}

		// 添加到结果
		result = append(result, txMap)
	}

	log.Infof("已获取地址 %s 的交易历史, 共 %d 条记录", address, len(result))
	return result, nil
}

// SendTransaction 发送交易示例
func SendTransaction(rawTx string) (string, error) {
	// 记录开始日志
	log.Infof("开始广播交易...")

	// 广播交易
	txid, err := BroadcastTransaction(rawTx)
	if err != nil {
		log.Errorf("广播交易失败: %v", err)
		return "", err
	}

	log.Infof("交易已成功广播, 交易ID: %s", txid)
	return txid, nil
}

// AsyncGetTransactionDetails 异步获取交易详情示例
func AsyncGetTransactionDetails(ctx context.Context, txid string) <-chan map[string]interface{} {
	resultChan := make(chan map[string]interface{}, 1)

	go func() {
		defer close(resultChan)

		// 检查上下文是否已取消
		select {
		case <-ctx.Done():
			log.Warnf("获取交易详情被取消: %v", ctx.Err())
			return
		default:
			// 继续执行
		}

		// 获取交易详情
		tx, err := GetTransaction(txid)
		if err != nil {
			log.Errorf("获取交易详情失败: %s, 错误: %v", txid, err)
			return
		}

		// 发送结果
		resultChan <- tx
	}()

	return resultChan
}

// BatchGetTransactions 批量获取交易详情示例
func BatchGetTransactions(txids []string, timeout time.Duration) []map[string]interface{} {
	// 创建上下文
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// 结果通道
	resultChans := make([]<-chan AsyncResult, len(txids))

	// 发起所有请求
	for i, txid := range txids {
		// 调用异步RPC方法
		resultChans[i] = CallMethodAsync(ctx, "blockchain.transaction.get", []interface{}{txid, true})
	}

	// 收集结果
	results := make([]map[string]interface{}, 0, len(txids))

	// 等待所有结果
	for i, resultChan := range resultChans {
		select {
		case result := <-resultChan:
			if result.Error != nil {
				log.Errorf("获取交易 %s 失败: %v", txids[i], result.Error)
				continue
			}

			// 解析交易
			var tx map[string]interface{}
			if err := json.Unmarshal(result.Result, &tx); err != nil {
				log.Errorf("解析交易 %s 失败: %v", txids[i], err)
				continue
			}

			// 添加到结果
			results = append(results, tx)

		case <-ctx.Done():
			log.Warnf("批量获取交易超时: %v", ctx.Err())
			return results
		}
	}

	log.Infof("批量获取交易完成, 共获取 %d/%d 个交易", len(results), len(txids))
	return results
}
