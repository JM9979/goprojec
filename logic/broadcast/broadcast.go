package logic

import (
	"context"
	"fmt"
	"net/http"

	"ginproject/entity/broadcast"
	"ginproject/middleware/log"
	"ginproject/repo/rpc/blockchain"
)

// BroadcastTxRaw 广播单个原始交易的业务逻辑
// 处理请求参数的验证、调用底层RPC接口以及处理响应
func BroadcastTxRaw(ctx context.Context, req *broadcast.TxBroadcastRequest) (*broadcast.BroadcastResponse, int, error) {
	// 验证请求参数
	if err := req.Validate(); err != nil {
		log.ErrorWithContext(ctx, "交易参数无效", "error", err)
		return nil, http.StatusBadRequest, err
	}

	// 记录API调用
	log.InfoWithContext(ctx, "开始广播原始交易", "txHexLength", len(req.TxHex))

	// 调用RPC广播交易
	resultChan := blockchain.SendRawTransaction(ctx, req.TxHex, false, false)
	result := <-resultChan

	// 处理错误
	if result.Error != nil {
		log.ErrorWithContext(ctx, "交易广播服务错误", "error", result.Error)
		return nil, http.StatusInternalServerError, result.Error
	}

	// 转换结果
	resp, ok := result.Result.(*broadcast.BroadcastResponse)
	if !ok {
		// 尝试将结果转换为字符串（txid）
		txid, ok := result.Result.(string)
		if !ok {
			err := ctx.Err()
			if err == nil {
				err = fmt.Errorf("结果类型转换失败: 期望 string 或 *broadcast.BroadcastResponse 类型")
			}
			log.ErrorWithContext(ctx, "结果类型转换失败", "type", fmt.Sprintf("%T", result.Result))
			return nil, http.StatusInternalServerError, err
		}
		// 如果是字符串，创建响应对象
		resp = &broadcast.BroadcastResponse{
			Result: txid,
		}
	}

	// 返回结果
	log.InfoWithContext(ctx, "交易广播请求处理完成",
		"success", true,
		"result", resp.Result)

	return resp, http.StatusOK, nil
}

// BroadcastTxsRaw 批量广播原始交易的业务逻辑
// 处理请求参数的验证、调用底层RPC接口以及处理响应
func BroadcastTxsRaw(ctx context.Context, req broadcast.TxsBroadcastRequest) (*broadcast.TxsBroadcastResponse, int, error) {
	// 验证请求参数
	if err := req.Validate(); err != nil {
		log.ErrorWithContext(ctx, "批量交易参数无效", "error", err)
		return nil, http.StatusBadRequest, err
	}

	// 记录API调用
	log.InfoWithContext(ctx, "开始批量广播原始交易", "count", len(req))

	// 准备发送的交易列表
	txList := make([]map[string]interface{}, 0, len(req))
	for _, txReq := range req {
		txList = append(txList, map[string]interface{}{
			"hex":           txReq.TxHex,
			"allowhighfees": false,
			"dontcheckfee":  false,
		})
	}

	// 调用RPC批量广播交易
	resultChan := blockchain.SendRawTransactions(ctx, txList)
	result := <-resultChan

	// 处理错误
	if result.Error != nil {
		log.ErrorWithContext(ctx, "批量交易广播服务错误", "error", result.Error)
		return nil, http.StatusInternalServerError, result.Error
	}

	// 转换结果
	resp, ok := result.Result.(*broadcast.TxsBroadcastResponse)
	if !ok {
		// 尝试处理不同类型的结果
		if mapResult, ok := result.Result.(map[string]interface{}); ok {
			// 创建响应对象
			resp = &broadcast.TxsBroadcastResponse{
				Result: &broadcast.TxsBroadcastResult{
					Invalid: []broadcast.InvalidTx{},
				},
			}

			// 处理可能的错误信息
			if errMap, ok := mapResult["error"].(map[string]interface{}); ok {
				code, _ := errMap["code"].(float64)
				message, _ := errMap["message"].(string)
				resp.Error = &broadcast.BroadcastError{
					Code:    int(code),
					Message: message,
				}
			}

			// 处理结果数据
			if resultMap, ok := mapResult["result"].(map[string]interface{}); ok {
				if invalidTxs, ok := resultMap["invalid"].([]interface{}); ok {
					invalid := make([]broadcast.InvalidTx, 0, len(invalidTxs))
					for _, inv := range invalidTxs {
						if invMap, ok := inv.(map[string]interface{}); ok {
							invalidTx := broadcast.InvalidTx{}
							if txid, ok := invMap["txid"].(string); ok {
								invalidTx.TxID = txid
							}
							if rejectCode, ok := invMap["reject_code"].(float64); ok {
								invalidTx.RejectCode = int(rejectCode)
							}
							if rejectReason, ok := invMap["reject_reason"].(string); ok {
								invalidTx.RejectReason = rejectReason
							}
							invalid = append(invalid, invalidTx)
						}
					}
					resp.Result.Invalid = invalid
				}
			}
		} else if _, ok := result.Result.(string); ok {
			// 如果是字符串，创建响应对象
			resp = &broadcast.TxsBroadcastResponse{
				Result: &broadcast.TxsBroadcastResult{
					Invalid: []broadcast.InvalidTx{},
				},
			}
		} else {
			err := ctx.Err()
			if err == nil {
				err = fmt.Errorf("结果类型转换失败: 期望 map、string 或 *broadcast.TxsBroadcastResponse 类型")
			}
			log.ErrorWithContext(ctx, "结果类型转换失败", "type", fmt.Sprintf("%T", result.Result))
			return nil, http.StatusInternalServerError, err
		}
	}

	// 如果有错误，返回对应的错误状态码
	statusCode := http.StatusOK
	if resp.Error != nil {
		log.WarnWithContext(ctx, "批量交易广播返回错误",
			"code", resp.Error.Code,
			"message", resp.Error.Message)
		statusCode = http.StatusBadRequest
	} else {
		invalidCount := 0
		if resp.Result != nil {
			invalidCount = len(resp.Result.Invalid)
		}
		log.InfoWithContext(ctx, "批量交易广播完成",
			"totalCount", len(req),
			"invalidCount", invalidCount)
	}
	// fmt.Println(resp.Result.TxID)
	// 返回结果
	return resp, statusCode, nil
}
