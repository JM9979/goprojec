package electrumx

import (
	"context"
	"encoding/json"
	"fmt"

	"ginproject/entity/electrumx"
	utility "ginproject/entity/utility"
	"ginproject/middleware/log"

	"go.uber.org/zap"
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
func GetBalance(scriptHash string) (*electrumx.BalanceResponse, error) {
	// 参数校验
	if len(scriptHash) == 0 {
		log.Errorf("调用GetBalance失败: scriptHash不能为空")
		return nil, fmt.Errorf("scriptHash不能为空")
	}

	// 记录开始调用日志
	log.Infof("开始获取脚本哈希余额: %s", scriptHash)

	// 调用RPC方法
	result, err := CallMethod("blockchain.scripthash.get_balance", []interface{}{scriptHash})
	if err != nil {
		log.Errorf("获取脚本哈希余额失败: %v", err)
		return nil, fmt.Errorf("获取脚本哈希余额失败: %w", err)
	}

	// 解析响应
	var balance electrumx.BalanceResponse
	if err := json.Unmarshal(result, &balance); err != nil {
		log.Errorf("解析脚本哈希余额失败: %v", err)
		return nil, fmt.Errorf("解析脚本哈希余额失败: %w", err)
	}

	log.Infof("成功获取脚本哈希余额: 已确认=%d, 未确认=%d", balance.Confirmed, balance.Unconfirmed)
	return &balance, nil
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

// GetListUnspent 获取指定脚本哈希的未花费交易输出
func GetListUnspent(scriptHash string) (electrumx.UtxoResponse, error) {
	// 参数校验
	if len(scriptHash) == 0 {
		log.Errorf("调用GetListUnspent失败: scriptHash不能为空")
		return nil, fmt.Errorf("scriptHash不能为空")
	}

	// 记录开始调用日志
	log.Infof("开始获取脚本哈希的UTXO: %s", scriptHash)

	// 调用RPC方法
	result, err := CallMethod("blockchain.scripthash.listunspent", []interface{}{scriptHash})
	if err != nil {
		log.Errorf("获取UTXO失败: %v", err)
		return nil, fmt.Errorf("获取UTXO失败: %w", err)
	}

	// 解析响应
	var utxos electrumx.UtxoResponse
	if err := json.Unmarshal(result, &utxos); err != nil {
		log.Errorf("解析UTXO响应失败: %v", err)
		return nil, fmt.Errorf("解析UTXO响应失败: %w", err)
	}

	log.Infof("成功获取UTXO, 数量: %d", len(utxos))
	return utxos, nil
}

// AsyncUtxoResult 异步UTXO结果
type AsyncUtxoResult struct {
	Result electrumx.UtxoResponse
	Error  error
}

// GetListUnspentAsync 异步获取指定脚本哈希的未花费交易输出
func GetListUnspentAsync(ctx context.Context, scriptHash string) <-chan AsyncUtxoResult {
	resultChan := make(chan AsyncUtxoResult, 1)

	go func() {
		defer close(resultChan)

		// 检查上下文是否已取消
		select {
		case <-ctx.Done():
			resultChan <- AsyncUtxoResult{
				Result: nil,
				Error:  ctx.Err(),
			}
			return
		default:
			// 继续执行
		}

		result, err := GetListUnspent(scriptHash)
		resultChan <- AsyncUtxoResult{
			Result: result,
			Error:  err,
		}
	}()

	return resultChan
}

// AddressToScriptHash 将比特币地址转换为脚本哈希
func AddressToScriptHash(address string) (string, error) {
	// 调用工具函数进行转换
	scriptHash, err := utility.AddressToScriptHash(address)
	if err != nil {
		log.Errorf("地址转换为脚本哈希失败: %v", err)
		return "", fmt.Errorf("地址转换为脚本哈希失败: %w", err)
	}

	return scriptHash, nil
}

// 错误定义
var (
	ErrEmptyScriptHash = fmt.Errorf("脚本哈希不能为空")
)

// GetUnspent 获取指定脚本哈希的未花费交易输出
func GetUnspent(ctx context.Context, scriptHash string) (electrumx.UtxoResponse, error) {
	log.InfoWithContext(ctx, "开始获取脚本哈希的UTXO", zap.String("scriptHash", scriptHash))

	// 参数校验
	if scriptHash == "" {
		log.ErrorWithContext(ctx, "脚本哈希不能为空")
		return nil, ErrEmptyScriptHash
	}

	// 调用ElectrumX RPC获取UTXO列表
	utxos, err := GetListUnspent(scriptHash)
	if err != nil {
		log.ErrorWithContext(ctx, "获取UTXO失败",
			zap.String("scriptHash", scriptHash),
			zap.Error(err))
		return nil, err
	}

	log.InfoWithContext(ctx, "成功获取UTXO",
		zap.String("scriptHash", scriptHash),
		zap.Int("count", len(utxos)))
	return utxos, nil
}

// ConvertAddressToScript 将WIF地址转换为脚本哈希
func ConvertAddressToScript(address string) (string, error) {
	// 调用AddressToScriptHash函数进行转换
	scriptHash, err := AddressToScriptHash(address)
	if err != nil {
		log.Errorf("地址转换为脚本哈希失败: %v", err)
		return "", err
	}

	return scriptHash, nil
}

// CallElectrumXRPC 通用的ElectrumX RPC调用函数，支持上下文控制
func CallElectrumXRPC(ctx context.Context, method string, params []interface{}) (json.RawMessage, error) {
	// 创建结果通道
	resultChan := make(chan struct {
		result json.RawMessage
		err    error
	}, 1)

	// 在goroutine中执行调用，避免阻塞
	go func() {
		// 记录开始调用日志
		log.InfoWithContext(ctx, "开始调用ElectrumX RPC",
			zap.String("method", method),
			zap.Any("params", params))

		// 获取客户端
		client, err := GetDefaultClient()
		if err != nil {
			log.ErrorWithContext(ctx, "获取ElectrumX客户端失败",
				zap.String("method", method),
				zap.Error(err))
			resultChan <- struct {
				result json.RawMessage
				err    error
			}{nil, fmt.Errorf("获取ElectrumX客户端失败: %w", err)}
			return
		}

		// 执行RPC调用
		result, err := client.CallRPC(method, params)

		// 发送结果到通道
		resultChan <- struct {
			result json.RawMessage
			err    error
		}{result, err}
	}()

	// 等待结果或上下文取消
	select {
	case <-ctx.Done():
		// 上下文已取消（超时或其他原因）
		log.WarnWithContext(ctx, "ElectrumX RPC调用已取消",
			zap.String("method", method),
			zap.Error(ctx.Err()))
		return nil, ctx.Err()
	case result := <-resultChan:
		// 收到结果
		if result.err != nil {
			log.ErrorWithContext(ctx, "ElectrumX RPC调用失败",
				zap.String("method", method),
				zap.Error(result.err))
			return nil, result.err
		}

		// 记录成功调用
		log.InfoWithContext(ctx, "ElectrumX RPC调用成功",
			zap.String("method", method),
			zap.Int("responseSize", len(result.result)))
		return result.result, nil
	}
}

// GetScriptHashHistoryAsync2 使用增强的方式异步获取指定脚本哈希的交易历史
func GetScriptHashHistoryAsync2(ctx context.Context, scriptHash string) (electrumx.ElectrumXHistoryResponse, error) {
	// 参数校验
	if scriptHash == "" {
		log.ErrorWithContext(ctx, "脚本哈希不能为空")
		return nil, ErrEmptyScriptHash
	}

	// 记录调用开始
	log.InfoWithContext(ctx, "开始获取脚本哈希历史",
		zap.String("scriptHash", scriptHash))

	// 通过新的通用调用函数执行RPC请求
	result, err := CallElectrumXRPC(ctx, "blockchain.scripthash.get_history", []interface{}{scriptHash})
	if err != nil {
		log.ErrorWithContext(ctx, "获取脚本哈希历史失败",
			zap.String("scriptHash", scriptHash),
			zap.Error(err))
		return nil, fmt.Errorf("获取脚本哈希历史失败: %w", err)
	}

	// 解析响应数据
	var history electrumx.ElectrumXHistoryResponse
	if err := json.Unmarshal(result, &history); err != nil {
		log.ErrorWithContext(ctx, "解析脚本哈希历史失败",
			zap.String("scriptHash", scriptHash),
			zap.Error(err))
		return nil, fmt.Errorf("解析脚本哈希历史失败: %w", err)
	}

	// 记录成功获取
	log.InfoWithContext(ctx, "成功获取脚本哈希历史",
		zap.String("scriptHash", scriptHash),
		zap.Int("recordCount", len(history)))

	return history, nil
}

// GetScriptHashBalance 获取指定脚本哈希的余额
func GetScriptHashBalance(scriptHash string) (*electrumx.BalanceResponse, error) {
	// 参数校验
	if len(scriptHash) == 0 {
		log.Errorf("调用GetScriptHashBalance失败: scriptHash不能为空")
		return nil, fmt.Errorf("scriptHash不能为空")
	}

	// 记录开始调用日志
	log.Infof("开始获取脚本哈希余额: %s", scriptHash)

	// 调用RPC方法
	result, err := CallMethod("blockchain.scripthash.get_balance", []interface{}{scriptHash})
	if err != nil {
		log.Errorf("获取脚本哈希余额失败: %v", err)
		return nil, fmt.Errorf("获取脚本哈希余额失败: %w", err)
	}

	// 解析响应
	var balance electrumx.BalanceResponse
	if err := json.Unmarshal(result, &balance); err != nil {
		log.Errorf("解析脚本哈希余额失败: %v, 原始数据: %s", err, string(result))
		return nil, fmt.Errorf("解析脚本哈希余额失败: %w", err)
	}

	log.Infof("成功获取脚本哈希余额: 已确认=%d, 未确认=%d", balance.Confirmed, balance.Unconfirmed)
	return &balance, nil
}

// GetScriptHashBalanceAsync 异步获取指定脚本哈希的余额
func GetScriptHashBalanceAsync(ctx context.Context, scriptHash string) <-chan AsyncBalanceResult {
	resultChan := make(chan AsyncBalanceResult, 1)

	go func() {
		defer close(resultChan)

		// 检查上下文是否已取消
		select {
		case <-ctx.Done():
			resultChan <- AsyncBalanceResult{
				Result: nil,
				Error:  ctx.Err(),
			}
			return
		default:
			// 继续执行
		}

		balance, err := GetScriptHashBalance(scriptHash)
		resultChan <- AsyncBalanceResult{
			Result: balance,
			Error:  err,
		}
	}()

	return resultChan
}

// AsyncBalanceResult 表示异步获取余额的结果
type AsyncBalanceResult struct {
	Result *electrumx.BalanceResponse
	Error  error
}

// GetAddressBalance 获取比特币地址的余额
func GetAddressBalance(ctx context.Context, address string) (*electrumx.AddressBalanceResponse, error) {
	// 参数校验
	if address == "" {
		log.ErrorWithContext(ctx, "地址不能为空")
		return nil, fmt.Errorf("地址不能为空")
	}

	// 记录调用开始
	log.InfoWithContext(ctx, "开始获取地址余额",
		zap.String("address", address))

	// 将地址转换为脚本哈希
	scriptHash, err := AddressToScriptHash(address)
	if err != nil {
		log.ErrorWithContext(ctx, "地址转换为脚本哈希失败",
			zap.String("address", address),
			zap.Error(err))
		return nil, fmt.Errorf("地址转换为脚本哈希失败: %w", err)
	}

	// 获取脚本哈希余额
	balance, err := GetScriptHashBalance(scriptHash)
	if err != nil {
		log.ErrorWithContext(ctx, "获取地址余额失败",
			zap.String("address", address),
			zap.String("scriptHash", scriptHash),
			zap.Error(err))
		return nil, fmt.Errorf("获取地址余额失败: %w", err)
	}

	// 计算总余额
	totalBalance := balance.Confirmed + balance.Unconfirmed

	// 构造响应
	response := &electrumx.AddressBalanceResponse{
		Balance:     totalBalance,
		Confirmed:   balance.Confirmed,
		Unconfirmed: balance.Unconfirmed,
	}

	log.InfoWithContext(ctx, "成功获取地址余额",
		zap.String("address", address),
		zap.Int64("balance", totalBalance),
		zap.Int64("confirmed", balance.Confirmed),
		zap.Int64("unconfirmed", balance.Unconfirmed))

	return response, nil
}

// GetScriptHashFrozenBalance 获取指定脚本哈希的冻结余额
func GetScriptHashFrozenBalance(scriptHash string) (*electrumx.FrozenBalanceResponse, error) {
	// 参数校验
	if len(scriptHash) == 0 {
		log.Errorf("调用GetScriptHashFrozenBalance失败: scriptHash不能为空")
		return nil, fmt.Errorf("scriptHash不能为空")
	}

	// 记录开始调用日志
	log.Infof("开始获取脚本哈希冻结余额: %s", scriptHash)

	// 调用RPC方法
	result, err := CallMethod("blockchain.scripthash.get_frozen_balance", []interface{}{scriptHash})
	if err != nil {
		log.Errorf("获取脚本哈希冻结余额失败: %v", err)
		return nil, fmt.Errorf("获取脚本哈希冻结余额失败: %w", err)
	}

	// 解析响应
	var frozenBalance electrumx.FrozenBalanceResponse
	if err := json.Unmarshal(result, &frozenBalance); err != nil {
		log.Errorf("解析脚本哈希冻结余额失败: %v, 原始数据: %s", err, string(result))
		return nil, fmt.Errorf("解析脚本哈希冻结余额失败: %w", err)
	}

	log.Infof("成功获取脚本哈希冻结余额: 冻结=%d", frozenBalance.Frozen)
	return &frozenBalance, nil
}

// GetAddressFrozenBalance 获取比特币地址的冻结余额
func GetAddressFrozenBalance(ctx context.Context, address string) (*electrumx.FrozenBalanceResponse, error) {
	// 参数校验
	if address == "" {
		log.ErrorWithContext(ctx, "地址不能为空")
		return nil, fmt.Errorf("地址不能为空")
	}

	// 记录调用开始
	log.InfoWithContext(ctx, "开始获取地址冻结余额",
		zap.String("address", address))

	// 将地址转换为脚本哈希
	scriptHash, err := AddressToScriptHash(address)
	if err != nil {
		log.ErrorWithContext(ctx, "地址转换为脚本哈希失败",
			zap.String("address", address),
			zap.Error(err))
		return nil, fmt.Errorf("地址转换为脚本哈希失败: %w", err)
	}

	// 获取脚本哈希冻结余额
	frozenBalance, err := GetScriptHashFrozenBalance(scriptHash)
	if err != nil {
		log.ErrorWithContext(ctx, "获取地址冻结余额失败",
			zap.String("address", address),
			zap.String("scriptHash", scriptHash),
			zap.Error(err))
		return nil, fmt.Errorf("获取地址冻结余额失败: %w", err)
	}

	log.InfoWithContext(ctx, "成功获取地址冻结余额",
		zap.String("address", address),
		zap.Int64("frozen", frozenBalance.Frozen))

	return frozenBalance, nil
}
