package electrumx

import (
	"context"
	"encoding/json"
	"fmt"

	"ginproject/entity/electrumx"
	"ginproject/middleware/log"
)

// 以下是使用连接池的高频方法实现

// CallMethodWithPool 通过连接池调用ElectrumX RPC方法的便捷函数
func CallMethodWithPool(ctx context.Context, method string, params []interface{}) (json.RawMessage, error) {
	// 记录开始调用日志
	log.InfoWithContext(ctx, "开始通过连接池调用ElectrumX方法:", method)

	result, err := CallPoolRPC(ctx, method, params)
	if err != nil {
		log.ErrorWithContext(ctx, "通过连接池调用ElectrumX方法失败:", method, "错误:", err)
		return nil, err
	}

	return result, nil
}

// GetBalanceWithPool 通过连接池获取地址余额
func GetBalanceWithPool(ctx context.Context, scriptHash string) (*electrumx.BalanceResponse, error) {
	// 参数校验
	if len(scriptHash) == 0 {
		log.ErrorWithContext(ctx, "调用GetBalanceWithPool失败: scriptHash不能为空")
		return nil, fmt.Errorf("scriptHash不能为空")
	}

	// 记录开始调用日志
	log.InfoWithContext(ctx, "开始通过连接池获取脚本哈希余额:", scriptHash)

	// 调用RPC方法
	result, err := CallMethodWithPool(ctx, "blockchain.scripthash.get_balance", []interface{}{scriptHash})
	if err != nil {
		log.ErrorWithContext(ctx, "通过连接池获取脚本哈希余额失败:", err)
		return nil, fmt.Errorf("获取脚本哈希余额失败: %w", err)
	}

	// 解析响应
	var balance electrumx.BalanceResponse
	if err := json.Unmarshal(result, &balance); err != nil {
		log.ErrorWithContext(ctx, "解析脚本哈希余额失败:", err, "原始数据:", string(result))
		return nil, fmt.Errorf("解析脚本哈希余额失败: %w", err)
	}

	log.InfoWithContext(ctx, "成功通过连接池获取脚本哈希余额: 已确认=", balance.Confirmed, "未确认=", balance.Unconfirmed)
	return &balance, nil
}

// GetListUnspentWithPool 通过连接池获取未花费交易输出
func GetListUnspentWithPool(ctx context.Context, scriptHash string) (electrumx.UtxoResponse, error) {
	// 参数校验
	if len(scriptHash) == 0 {
		log.ErrorWithContext(ctx, "调用GetListUnspentWithPool失败: scriptHash不能为空")
		return nil, fmt.Errorf("scriptHash不能为空")
	}

	// 记录开始调用日志
	log.InfoWithContext(ctx, "开始通过连接池获取脚本哈希未花费交易输出:", scriptHash)

	// 调用RPC方法
	result, err := CallMethodWithPool(ctx, "blockchain.scripthash.listunspent", []interface{}{scriptHash})
	if err != nil {
		log.ErrorWithContext(ctx, "通过连接池获取脚本哈希未花费交易输出失败:", err)
		return nil, fmt.Errorf("获取脚本哈希未花费交易输出失败: %w", err)
	}

	// 解析响应
	var utxos electrumx.UtxoResponse
	if err := json.Unmarshal(result, &utxos); err != nil {
		log.ErrorWithContext(ctx, "解析脚本哈希未花费交易输出失败:", err, "原始数据:", string(result))
		return nil, fmt.Errorf("解析脚本哈希未花费交易输出失败: %w", err)
	}

	log.InfoWithContext(ctx, "成功通过连接池获取脚本哈希未花费交易输出, 数量:", len(utxos))
	return utxos, nil
}

// GetScriptHashHistoryWithPool 通过连接池获取脚本哈希交易历史
func GetScriptHashHistoryWithPool(ctx context.Context, scriptHash string) (electrumx.ElectrumXHistoryResponse, error) {
	// 参数校验
	if len(scriptHash) == 0 {
		log.ErrorWithContext(ctx, "调用GetScriptHashHistoryWithPool失败: scriptHash不能为空")
		return nil, fmt.Errorf("scriptHash不能为空")
	}

	// 记录开始调用日志
	log.InfoWithContext(ctx, "开始通过连接池获取脚本哈希交易历史:", scriptHash)

	// 调用RPC方法
	result, err := CallMethodWithPool(ctx, "blockchain.scripthash.get_history", []interface{}{scriptHash})
	if err != nil {
		log.ErrorWithContext(ctx, "通过连接池获取脚本哈希交易历史失败:", err)
		return nil, fmt.Errorf("获取脚本哈希交易历史失败: %w", err)
	}

	// 解析响应
	var history electrumx.ElectrumXHistoryResponse
	if err := json.Unmarshal(result, &history); err != nil {
		log.ErrorWithContext(ctx, "解析脚本哈希交易历史失败:", err, "原始数据:", string(result))
		return nil, fmt.Errorf("解析脚本哈希交易历史失败: %w", err)
	}

	log.InfoWithContext(ctx, "成功通过连接池获取脚本哈希交易历史, 记录数:", len(history))
	return history, nil
}

// BroadcastTransactionWithPool 通过连接池广播交易
func BroadcastTransactionWithPool(ctx context.Context, rawTx string) (string, error) {
	// 参数校验
	if len(rawTx) == 0 {
		log.ErrorWithContext(ctx, "调用BroadcastTransactionWithPool失败: rawTx不能为空")
		return "", fmt.Errorf("rawTx不能为空")
	}

	// 记录开始调用日志
	log.InfoWithContext(ctx, "开始通过连接池广播交易")

	// 调用RPC方法
	result, err := CallMethodWithPool(ctx, "blockchain.transaction.broadcast", []interface{}{rawTx})
	if err != nil {
		log.ErrorWithContext(ctx, "通过连接池广播交易失败:", err)
		return "", fmt.Errorf("广播交易失败: %w", err)
	}

	// 解析响应
	var txid string
	if err := json.Unmarshal(result, &txid); err != nil {
		log.ErrorWithContext(ctx, "解析交易ID失败:", err, "原始数据:", string(result))
		return "", fmt.Errorf("解析交易ID失败: %w", err)
	}

	log.InfoWithContext(ctx, "成功通过连接池广播交易, 交易ID:", txid)
	return txid, nil
}

// GetTransactionWithPool 通过连接池获取交易详情
func GetTransactionWithPool(ctx context.Context, txid string, verbose bool) (json.RawMessage, error) {
	// 参数校验
	if len(txid) == 0 {
		log.ErrorWithContext(ctx, "调用GetTransactionWithPool失败: txid不能为空")
		return nil, fmt.Errorf("txid不能为空")
	}

	// 记录开始调用日志
	log.InfoWithContext(ctx, "开始通过连接池获取交易详情:", txid, "verbose:", verbose)

	// 调用RPC方法
	result, err := CallMethodWithPool(ctx, "blockchain.transaction.get", []interface{}{txid, verbose})
	if err != nil {
		log.ErrorWithContext(ctx, "通过连接池获取交易详情失败:", err)
		return nil, fmt.Errorf("获取交易详情失败: %w", err)
	}

	log.InfoWithContext(ctx, "成功通过连接池获取交易详情:", txid)
	return result, nil
}

// EstimateFeeWithPool 通过连接池估算交易手续费
func EstimateFeeWithPool(ctx context.Context, blocks int) (float64, error) {
	// 参数校验
	if blocks <= 0 {
		blocks = 1 // 默认值
	}

	// 记录开始调用日志
	log.InfoWithContext(ctx, "开始通过连接池估算交易手续费, blocks:", blocks)

	// 调用RPC方法
	result, err := CallMethodWithPool(ctx, "blockchain.estimatefee", []interface{}{blocks})
	if err != nil {
		log.ErrorWithContext(ctx, "通过连接池估算交易手续费失败:", err)
		return 0, fmt.Errorf("估算交易手续费失败: %w", err)
	}

	// 解析响应
	var fee float64
	if err := json.Unmarshal(result, &fee); err != nil {
		log.ErrorWithContext(ctx, "解析交易手续费失败:", err, "原始数据:", string(result))
		return 0, fmt.Errorf("解析交易手续费失败: %w", err)
	}

	log.InfoWithContext(ctx, "成功通过连接池估算交易手续费:", fee)
	return fee, nil
}

// GetAddressBalanceWithPool 通过连接池获取地址余额
func GetAddressBalanceWithPool(ctx context.Context, address string) (*electrumx.AddressBalanceResponse, error) {
	// 参数校验
	if len(address) == 0 {
		log.ErrorWithContext(ctx, "调用GetAddressBalanceWithPool失败: address不能为空")
		return nil, fmt.Errorf("address不能为空")
	}

	// 将地址转换为脚本哈希
	scriptHash, err := AddressToScriptHash(ctx, address)
	if err != nil {
		log.ErrorWithContext(ctx, "地址转换为脚本哈希失败:", err)
		return nil, fmt.Errorf("地址转换为脚本哈希失败: %w", err)
	}

	// 获取余额
	balance, err := GetBalanceWithPool(ctx, scriptHash)
	if err != nil {
		return nil, err
	}

	// 构造响应
	addressBalance := &electrumx.AddressBalanceResponse{
		Balance:     balance.Confirmed + balance.Unconfirmed,
		Confirmed:   balance.Confirmed,
		Unconfirmed: balance.Unconfirmed,
	}

	return addressBalance, nil
}

// GetPoolStats 获取连接池统计信息
func GetPoolStats(ctx context.Context) (idleConns, openConns int, err error) {
	pool, err := GetClientPool()
	if err != nil {
		log.ErrorWithContext(ctx, "获取ElectrumX客户端池失败:", err)
		return 0, 0, err
	}

	idleConns, openConns = pool.Stats()
	log.InfoWithContext(ctx, "ElectrumX连接池状态: 空闲连接数=", idleConns, ", 打开连接数=", openConns)
	return
}
