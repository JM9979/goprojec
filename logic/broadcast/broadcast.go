package logic

import (
	"context"
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
		err := ctx.Err()
		if err == nil {
			err = ctx.Err() // 尝试获取上下文错误
		}
		log.ErrorWithContext(ctx, "结果类型转换失败", "type", "BroadcastResponse")
		return nil, http.StatusInternalServerError, err
	}

	// 返回结果
	log.InfoWithContext(ctx, "交易广播请求处理完成",
		"success", resp.Error == nil,
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
		err := ctx.Err()
		if err == nil {
			err = ctx.Err() // 尝试获取上下文错误
		}
		log.ErrorWithContext(ctx, "结果类型转换失败", "type", "TxsBroadcastResponse")
		return nil, http.StatusInternalServerError, err
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

	// 返回结果
	return resp, statusCode, nil
}
