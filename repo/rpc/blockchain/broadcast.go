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

		// 调用RPC
		result, err := CallRPC("sendrawtransaction", []interface{}{txHex, allowHighFees, bypassLimits}, false)
		if err != nil {
			log.ErrorWithContext(ctx, "广播原始交易失败", "error", err)
			resultChan <- AsyncResult{
				Result: nil,
				Error:  fmt.Errorf("广播原始交易失败: %w", err),
			}
			return
		}

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
		log.InfoWithContext(ctx, "开始批量广播原始交易(完整响应)", "count", len(txList))

		// 调用RPC
		result, err := CallRPC("sendrawtransactions", []interface{}{txList}, true)
		if err != nil {
			log.ErrorWithContext(ctx, "批量广播原始交易失败", "error", err)
			resultChan <- AsyncResult{
				Result: &broadcast.TxsBroadcastResponse{
					Error: &broadcast.BroadcastError{
						Code:    -1,
						Message: err.Error(),
					},
				},
				Error: nil,
			}
			return
		}

		// 解析完整响应
		rpcResp, ok := result.(RPCResponse)
		if !ok {
			log.ErrorWithContext(ctx, "解析RPC响应失败", "result", result)
			resultChan <- AsyncResult{
				Result: &broadcast.TxsBroadcastResponse{
					Error: &broadcast.BroadcastError{
						Code:    -1,
						Message: "解析RPC响应失败",
					},
				},
				Error: nil,
			}
			return
		}

		// 处理错误情况
		if rpcResp.Error != nil {
			log.WarnWithContext(ctx, "批量广播交易返回错误",
				"code", rpcResp.Error.Code,
				"message", rpcResp.Error.Message)

			resultChan <- AsyncResult{
				Result: &broadcast.TxsBroadcastResponse{
					Error: &broadcast.BroadcastError{
						Code:    rpcResp.Error.Code,
						Message: rpcResp.Error.Message,
					},
				},
				Error: nil,
			}
			return
		}

		// 解析结果
		var resultMap map[string]interface{}
		resultBytes, ok := rpcResp.Result.([]byte)
		if !ok {
			// 尝试转换为json原始消息
			rawMsg, ok := rpcResp.Result.(json.RawMessage)
			if !ok {
				log.ErrorWithContext(ctx, "解析批量广播结果失败", "result", fmt.Sprintf("%T", rpcResp.Result))
				resultChan <- AsyncResult{
					Result: &broadcast.TxsBroadcastResponse{
						Error: &broadcast.BroadcastError{
							Code:    -1,
							Message: "解析批量广播结果失败: 无法转换结果类型",
						},
					},
					Error: nil,
				}
				return
			}
			resultBytes = []byte(rawMsg)
		}

		if err := json.Unmarshal(resultBytes, &resultMap); err != nil {
			log.ErrorWithContext(ctx, "解析批量广播结果失败", "result", string(resultBytes))
			resultChan <- AsyncResult{
				Result: &broadcast.TxsBroadcastResponse{
					Error: &broadcast.BroadcastError{
						Code:    -1,
						Message: "解析批量广播结果失败: " + err.Error(),
					},
				},
				Error: nil,
			}
			return
		}

		// 处理无效交易列表
		txsBroadcastResult := &broadcast.TxsBroadcastResult{
			Invalid: []broadcast.InvalidTx{},
		}

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
