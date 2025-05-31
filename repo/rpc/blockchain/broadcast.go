package blockchain

import (
	"context"
	"encoding/json"
	"fmt"

	"ginproject/entity/broadcast"
	"ginproject/middleware/log"
)

// SendRawTransaction 发送原始交易
func SendRawTransaction(ctx context.Context, txHex string, allowHighFees bool, bypassLimits bool) <-chan AsyncResult {
	resultChan := make(chan AsyncResult, 1)

	go func() {
		defer close(resultChan)

		// 记录开始调用日志
		log.InfoWithContext(ctx, "开始广播原始交易", "txHexLength", len(txHex))

		// 使用异步方式调用RPC
		asyncChan := CallRPCAsync(ctx, "sendrawtransaction", []interface{}{txHex, allowHighFees, bypassLimits}, false)
		asyncResult := <-asyncChan

		if asyncResult.Error != nil {
			log.ErrorWithContext(ctx, "广播原始交易失败", "error", asyncResult.Error)
			resultChan <- AsyncResult{
				Result: nil,
				Error:  fmt.Errorf("广播原始交易失败: %w", asyncResult.Error),
			}
			return
		}

		result := asyncResult.Result

		// 将结果转换为字符串
		txid, ok := result.(string)
		if !ok {
			log.ErrorWithContext(ctx, "解析交易ID失败", "result", result)
			resultChan <- AsyncResult{
				Result: nil,
				Error:  fmt.Errorf("解析交易ID失败"),
			}
			return
		}

		log.InfoWithContext(ctx, "成功广播原始交易", "txid", txid)
		resultChan <- AsyncResult{
			Result: txid,
			Error:  nil,
		}
	}()

	return resultChan
}

// SendRawTransactions 批量发送原始交易
func SendRawTransactions(ctx context.Context, txList []map[string]interface{}) <-chan AsyncResult {
	resultChan := make(chan AsyncResult, 1)

	go func() {
		defer close(resultChan)

		// 记录开始调用日志
		log.InfoWithContext(ctx, "开始批量广播原始交易", "count", len(txList))

		// 如果只有一笔交易，尝试使用单笔交易广播
		if len(txList) == 1 {
			log.InfoWithContext(ctx, "只有一笔交易，尝试使用单笔交易广播")
			tx := txList[0]
			singleChan := SendRawTransaction(ctx, tx["hex"].(string),
				tx["allowhighfees"].(bool),
				tx["dontcheckfee"].(bool))
			singleResult := <-singleChan

			if singleResult.Error != nil {
				log.ErrorWithContext(ctx, "单笔交易广播失败", "error", singleResult.Error)
				resultChan <- AsyncResult{
					Result: &broadcast.TxsBroadcastResponse{
						Error: &broadcast.BroadcastError{
							Code:    -1,
							Message: singleResult.Error.Error(),
						},
					},
					Error: nil,
				}
				return
			}

			// 处理单笔交易的结果
			if txid, ok := singleResult.Result.(string); ok {
				txsBroadcastResult := &broadcast.TxsBroadcastResult{
					TxIDs:   []string{txid},
					Invalid: []broadcast.InvalidTx{},
				}
				log.InfoWithContext(ctx, "单笔交易广播成功", "txid", txid)
				resultChan <- AsyncResult{
					Result: &broadcast.TxsBroadcastResponse{
						Result: txsBroadcastResult,
					},
					Error: nil,
				}
				return
			}
		}

		// 使用异步方式调用RPC
		asyncChan := CallRPCAsync(ctx, "sendrawtransactions", []interface{}{txList}, false)
		asyncResult := <-asyncChan

		if asyncResult.Error != nil {
			log.ErrorWithContext(ctx, "批量广播原始交易失败", "error", asyncResult.Error)
			resultChan <- AsyncResult{
				Result: &broadcast.TxsBroadcastResponse{
					Error: &broadcast.BroadcastError{
						Code:    -1,
						Message: asyncResult.Error.Error(),
					},
				},
				Error: nil,
			}
			return
		}

		result := asyncResult.Result

		// 解析完整响应
		var resultMap map[string]interface{}

		// 尝试直接转换为 map
		if m, ok := result.(map[string]interface{}); ok {
			resultMap = m
			// 打印完整的结果结构
			log.InfoWithContext(ctx, "批量广播返回结果", "result", resultMap)
		} else {
			// 尝试将结果转换为 json 并解析
			resultBytes, err := json.Marshal(result)
			if err != nil {
				log.ErrorWithContext(ctx, "序列化批量广播结果失败", "error", err)
				resultChan <- AsyncResult{
					Result: &broadcast.TxsBroadcastResponse{
						Error: &broadcast.BroadcastError{
							Code:    -1,
							Message: fmt.Sprintf("序列化批量广播结果失败: %v", err),
						},
					},
					Error: nil,
				}
				return
			}

			if err := json.Unmarshal(resultBytes, &resultMap); err != nil {
				log.ErrorWithContext(ctx, "解析批量广播结果失败", "error", err)
				resultChan <- AsyncResult{
					Result: &broadcast.TxsBroadcastResponse{
						Error: &broadcast.BroadcastError{
							Code:    -1,
							Message: fmt.Sprintf("解析批量广播结果失败: %v", err),
						},
					},
					Error: nil,
				}
				return
			}
			// 打印完整的结果结构
			log.InfoWithContext(ctx, "批量广播返回结果(JSON解析后)", "result", resultMap)
		}

		// 处理无效交易列表
		txsBroadcastResult := &broadcast.TxsBroadcastResult{
			TxIDs:   make([]string, 0),
			Invalid: []broadcast.InvalidTx{},
		}

		// 处理成功的交易ID列表
		if result, exists := resultMap["result"]; exists && result != nil {
			log.InfoWithContext(ctx, "处理 result 字段", "result", result)

			if resultMap, ok := result.(map[string]interface{}); ok {
				log.InfoWithContext(ctx, "result 是 map 类型", "resultMap", resultMap)

				// 尝试从 result.txids 获取
				if txids, exists := resultMap["txids"]; exists && txids != nil {
					log.InfoWithContext(ctx, "找到 txids 字段", "txids", txids)
					if txidArray, ok := txids.([]interface{}); ok {
						for _, txid := range txidArray {
							if txidStr, ok := txid.(string); ok {
								txsBroadcastResult.TxIDs = append(txsBroadcastResult.TxIDs, txidStr)
							}
						}
					}
				} else if txid, exists := resultMap["txid"]; exists && txid != nil {
					log.InfoWithContext(ctx, "找到 txid 字段", "txid", txid)
					// 尝试从 result.txid 获取
					if txidStr, ok := txid.(string); ok {
						txsBroadcastResult.TxIDs = append(txsBroadcastResult.TxIDs, txidStr)
					}
				}
			} else if txidArray, ok := result.([]interface{}); ok {
				// 直接尝试将 result 作为交易ID数组处理
				log.InfoWithContext(ctx, "result 是数组类型", "array", txidArray)
				for _, txid := range txidArray {
					if txidStr, ok := txid.(string); ok {
						txsBroadcastResult.TxIDs = append(txsBroadcastResult.TxIDs, txidStr)
					}
				}
			} else if txidStr, ok := result.(string); ok {
				// 直接尝试将 result 作为单个交易ID处理
				log.InfoWithContext(ctx, "result 是字符串类型", "txid", txidStr)
				txsBroadcastResult.TxIDs = append(txsBroadcastResult.TxIDs, txidStr)
			}
		}

		if len(txsBroadcastResult.TxIDs) == 0 {
			log.WarnWithContext(ctx, "未找到任何成功的交易ID", "resultMap", resultMap)
		}

		// 处理无效交易列表
		if invalidArray, exists := resultMap["invalid"]; exists && invalidArray != nil {
			if invalidItems, ok := invalidArray.([]interface{}); ok {
				for _, item := range invalidItems {
					if itemMap, ok := item.(map[string]interface{}); ok {
						invalid := broadcast.InvalidTx{}

						if txid, ok := itemMap["txid"].(string); ok {
							invalid.TxID = txid
						}

						if code, ok := itemMap["reject_code"].(float64); ok {
							invalid.RejectCode = int(code)
						}

						if reason, ok := itemMap["reject_reason"].(string); ok {
							invalid.RejectReason = reason
						}

						txsBroadcastResult.Invalid = append(txsBroadcastResult.Invalid, invalid)
					}
				}
			}
		}

		log.InfoWithContext(ctx, "成功批量广播原始交易",
			"total", len(txList),
			"success", len(txsBroadcastResult.TxIDs),
			"txids", txsBroadcastResult.TxIDs,
			"invalid", len(txsBroadcastResult.Invalid))

		resultChan <- AsyncResult{
			Result: &broadcast.TxsBroadcastResponse{
				Result: txsBroadcastResult,
			},
			Error: nil,
		}
	}()

	return resultChan
}
