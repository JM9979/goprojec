package transaction

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"ginproject/entity/transaction"
	"ginproject/middleware/log"
	"ginproject/repo/rpc/blockchain"
)

// BroadcastTxRaw 广播单个原始交易的业务逻辑
// 处理请求参数的验证、调用底层RPC接口以及处理响应
func BroadcastTxRaw(ctx context.Context, req *transaction.TxBroadcastRequest) (*transaction.BroadcastResponse, int, error) {
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

	// 转换结果为txid字符串
	txid, ok := result.Result.(string)
	if !ok {
		log.ErrorWithContext(ctx, "结果类型转换失败", "type", fmt.Sprintf("%T", result.Result))
		return nil, http.StatusInternalServerError, fmt.Errorf("结果类型转换失败")
	}

	// 创建响应
	resp := &transaction.BroadcastResponse{
		Result: txid,
	}

	// 返回结果
	log.InfoWithContext(ctx, "交易广播请求处理完成", "txid", txid)
	return resp, http.StatusOK, nil
}

// DecodeRawTx 解码原始交易的业务逻辑
func DecodeRawTx(ctx context.Context, req *transaction.TxDecodeRawRequest) (*transaction.TxDecodeResponse, int, error) {
	// 验证请求参数
	if err := req.Validate(); err != nil {
		log.ErrorWithContext(ctx, "解码交易参数无效", "error", err)
		return nil, http.StatusBadRequest, err
	}

	// 记录API调用
	log.InfoWithContext(ctx, "开始解码原始交易", "txHexLength", len(req.TxHex))

	// 调用RPC解码交易
	resultChan := blockchain.DecodeRawTransaction(ctx, req.TxHex)
	result := <-resultChan

	// 处理错误
	if result.Error != nil {
		log.ErrorWithContext(ctx, "解码交易服务错误", "error", result.Error)
		return nil, http.StatusInternalServerError, result.Error
	}

	// 转换结果
	jsonData, err := json.Marshal(result.Result)
	if err != nil {
		log.ErrorWithContext(ctx, "解码交易结果序列化失败", "error", err)
		return nil, http.StatusInternalServerError, fmt.Errorf("解码交易结果序列化失败: %w", err)
	}

	var resp transaction.TxDecodeResponse
	if err := json.Unmarshal(jsonData, &resp); err != nil {
		log.ErrorWithContext(ctx, "解码交易结果反序列化失败", "error", err)
		return nil, http.StatusInternalServerError, fmt.Errorf("解码交易结果反序列化失败: %w", err)
	}

	// 返回结果
	log.InfoWithContext(ctx, "解码原始交易完成", "txid", resp.TxID)
	return &resp, http.StatusOK, nil
}

// GetTxRawHex 获取交易原始十六进制数据的业务逻辑
func GetTxRawHex(ctx context.Context, txid string) (string, int, error) {
	// 验证参数
	if txid == "" {
		log.ErrorWithContext(ctx, "获取交易原始数据失败：交易ID不能为空")
		return "", http.StatusBadRequest, fmt.Errorf("交易ID不能为空")
	}

	// 记录API调用
	log.InfoWithContext(ctx, "开始获取交易原始数据", "txid", txid)

	// 调用RPC获取交易
	resultChan := blockchain.GetRawTransaction(ctx, txid, false)
	result := <-resultChan

	// 处理错误
	if result.Error != nil {
		log.ErrorWithContext(ctx, "获取交易原始数据服务错误", "error", result.Error)
		return "", http.StatusInternalServerError, result.Error
	}

	// 转换结果
	txHex, ok := result.Result.(string)
	if !ok {
		log.ErrorWithContext(ctx, "获取交易原始数据结果类型转换失败", "type", fmt.Sprintf("%T", result.Result))
		return "", http.StatusInternalServerError, fmt.Errorf("获取交易原始数据结果类型转换失败")
	}

	// 返回结果
	log.InfoWithContext(ctx, "获取交易原始数据完成", "txid", txid, "hexLength", len(txHex))
	return txHex, http.StatusOK, nil
}

// DecodeTxByHash 通过交易ID解码交易的业务逻辑
func DecodeTxByHash(ctx context.Context, txid string) (*transaction.TxDecodeResponse, int, error) {
	// 验证参数
	if txid == "" {
		log.ErrorWithContext(ctx, "解码交易失败：交易ID不能为空")
		return nil, http.StatusBadRequest, fmt.Errorf("交易ID不能为空")
	}

	// 记录API调用
	log.InfoWithContext(ctx, "开始通过交易ID解码交易", "txid", txid)

	// 调用RPC获取交易详情
	resultChan := blockchain.GetRawTransaction(ctx, txid, true)
	result := <-resultChan

	// 处理错误
	if result.Error != nil {
		log.ErrorWithContext(ctx, "解码交易服务错误", "error", result.Error)
		return nil, http.StatusInternalServerError, result.Error
	}

	// 转换结果
	jsonData, err := json.Marshal(result.Result)
	if err != nil {
		log.ErrorWithContext(ctx, "解码交易结果序列化失败", "error", err)
		return nil, http.StatusInternalServerError, fmt.Errorf("解码交易结果序列化失败: %w", err)
	}

	var resp transaction.TxDecodeResponse
	if err := json.Unmarshal(jsonData, &resp); err != nil {
		log.ErrorWithContext(ctx, "解码交易结果反序列化失败", "error", err)
		return nil, http.StatusInternalServerError, fmt.Errorf("解码交易结果反序列化失败: %w", err)
	}

	// 返回结果
	log.InfoWithContext(ctx, "通过交易ID解码交易完成", "txid", txid)
	return &resp, http.StatusOK, nil
}

// GetTxVins 获取交易输入数据的业务逻辑
func GetTxVins(ctx context.Context, txids []string) ([]transaction.TxVinsRawResponse, int, error) {
	// 验证参数
	if len(txids) == 0 {
		log.ErrorWithContext(ctx, "获取交易输入数据失败：交易ID列表不能为空")
		return nil, http.StatusBadRequest, fmt.Errorf("交易ID列表不能为空")
	}

	// 记录API调用
	log.InfoWithContext(ctx, "开始获取交易输入数据", "txids", txids)

	// 调用RPC获取交易输入数据
	resultChan := blockchain.GetTxVins(ctx, txids)
	result := <-resultChan

	// 处理错误
	if result.Error != nil {
		log.ErrorWithContext(ctx, "获取交易输入数据服务错误", "error", result.Error)
		return nil, http.StatusInternalServerError, result.Error
	}

	// 转换结果
	jsonData, err := json.Marshal(result.Result)
	if err != nil {
		log.ErrorWithContext(ctx, "获取交易输入数据结果序列化失败", "error", err)
		return nil, http.StatusInternalServerError, fmt.Errorf("获取交易输入数据结果序列化失败: %w", err)
	}

	var resp []transaction.TxVinsRawResponse
	if err := json.Unmarshal(jsonData, &resp); err != nil {
		log.ErrorWithContext(ctx, "获取交易输入数据结果反序列化失败", "error", err)
		return nil, http.StatusInternalServerError, fmt.Errorf("获取交易输入数据结果反序列化失败: %w", err)
	}

	// 返回结果
	log.InfoWithContext(ctx, "获取交易输入数据完成", "count", len(resp))
	return resp, http.StatusOK, nil
}
