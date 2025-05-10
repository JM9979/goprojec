package electrumx

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"strconv"

	"ginproject/entity/electrumx"
	utility "ginproject/entity/utility"
	"ginproject/middleware/log"
)

// Global client instance
var (
	defaultClient *ElectrumXClient
	initialized   = false
	poolEnabled   = true // 默认启用连接池
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

		// 默认启用连接池
		if poolEnabled {
			if err := defaultClient.EnablePool(); err != nil {
				log.Error("启用ElectrumX连接池失败:", err)
			} else {
				log.Info("ElectrumX连接池已启用")
			}
		}
	}

	return defaultClient, nil
}

// EnablePool 启用ElectrumX连接池
func EnablePool() error {
	if !initialized {
		// 如果客户端尚未初始化，设置标志位，客户端初始化时会启用连接池
		poolEnabled = true
		return nil
	}

	// 客户端已初始化，直接启用连接池
	return defaultClient.EnablePool()
}

// DisablePool 禁用ElectrumX连接池
func DisablePool() error {
	if !initialized {
		poolEnabled = false
		return nil
	}

	return defaultClient.DisablePool()
}

// GetClientPoolStats 获取连接池统计信息
func GetClientPoolStats() (idleConns, openConns int, err error) {
	if !initialized {
		return 0, 0, fmt.Errorf("ElectrumX客户端尚未初始化")
	}

	return defaultClient.PoolStats()
}

// CallMethod 调用ElectrumX RPC方法的简便函数
func CallMethod(ctx context.Context, method string, params []interface{}) (json.RawMessage, error) {
	client, err := GetDefaultClient()
	if err != nil {
		return nil, fmt.Errorf("获取ElectrumX客户端失败: %w", err)
	}

	// 记录开始调用日志
	log.InfoWithContext(ctx, "开始调用ElectrumX方法:", method)

	result, err := client.CallRPCWithContext(ctx, method, params)
	if err != nil {
		log.ErrorWithContext(ctx, "调用ElectrumX方法失败:", method, "错误:", err)
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

		result, err := CallMethod(ctx, method, params)
		resultChan <- AsyncResult{
			Result: result,
			Error:  err,
		}
	}()

	return resultChan
}

// 以下是常用的ElectrumX RPC方法

// GetBlockHeader 获取区块头信息
func GetBlockHeader(ctx context.Context, height int) (map[string]interface{}, error) {
	resultChan := CallMethodAsync(ctx, "blockchain.block.header", []interface{}{height})
	result := <-resultChan
	if result.Error != nil {
		return nil, result.Error
	}

	var header map[string]interface{}
	if err := json.Unmarshal(result.Result, &header); err != nil {
		log.ErrorWithContext(ctx, "解析区块头失败:", err)
		return nil, fmt.Errorf("解析区块头失败: %w", err)
	}

	return header, nil
}

// GetBalance 获取地址余额
func GetBalance(ctx context.Context, scriptHash string) (*electrumx.BalanceResponse, error) {
	// 参数校验
	if len(scriptHash) == 0 {
		log.ErrorWithContext(ctx, "调用GetBalance失败: scriptHash不能为空")
		return nil, fmt.Errorf("scriptHash不能为空")
	}

	// 记录开始调用日志
	log.InfoWithContext(ctx, "开始获取脚本哈希余额:", scriptHash)

	// 调用RPC方法（改为异步）
	resultChan := CallMethodAsync(ctx, "blockchain.scripthash.get_balance", []interface{}{scriptHash})
	result := <-resultChan
	if result.Error != nil {
		log.ErrorWithContext(ctx, "获取脚本哈希余额失败:", result.Error)
		return nil, fmt.Errorf("获取脚本哈希余额失败: %w", result.Error)
	}

	// 解析响应
	var balance electrumx.BalanceResponse
	if err := json.Unmarshal(result.Result, &balance); err != nil {
		log.ErrorWithContext(ctx, "解析脚本哈希余额失败:", err, "原始数据:", string(result.Result))
		return nil, fmt.Errorf("解析脚本哈希余额失败: %w", err)
	}

	log.InfoWithContext(ctx, "成功获取脚本哈希余额: 已确认=", balance.Confirmed, "未确认=", balance.Unconfirmed)
	return &balance, nil
}

// GetTransactionHistory 获取地址交易历史
func GetTransactionHistory(ctx context.Context, address string) ([]interface{}, error) {
	resultChan := CallMethodAsync(ctx, "blockchain.scripthash.get_history", []interface{}{address})
	result := <-resultChan
	if result.Error != nil {
		return nil, result.Error
	}

	var history []interface{}
	if err := json.Unmarshal(result.Result, &history); err != nil {
		log.ErrorWithContext(ctx, "解析交易历史失败:", err)
		return nil, fmt.Errorf("解析交易历史失败: %w", err)
	}

	return history, nil
}

// BroadcastTransaction 广播原始交易
func BroadcastTransaction(ctx context.Context, rawTx string) (string, error) {
	resultChan := CallMethodAsync(ctx, "blockchain.transaction.broadcast", []interface{}{rawTx})
	result := <-resultChan
	if result.Error != nil {
		return "", result.Error
	}

	var txid string
	if err := json.Unmarshal(result.Result, &txid); err != nil {
		log.ErrorWithContext(ctx, "解析交易ID失败:", err)
		return "", fmt.Errorf("解析交易ID失败: %w", err)
	}

	return txid, nil
}

// GetTransaction 获取交易详情
func GetTransaction(ctx context.Context, txid string) (map[string]interface{}, error) {
	resultChan := CallMethodAsync(ctx, "blockchain.transaction.get", []interface{}{txid, true})
	result := <-resultChan
	if result.Error != nil {
		return nil, result.Error
	}

	var tx map[string]interface{}
	if err := json.Unmarshal(result.Result, &tx); err != nil {
		log.ErrorWithContext(ctx, "解析交易详情失败:", err)
		return nil, fmt.Errorf("解析交易详情失败: %w", err)
	}

	return tx, nil
}

// EstimateFee 估算交易手续费
func EstimateFee(ctx context.Context, blocks int) (float64, error) {
	resultChan := CallMethodAsync(ctx, "blockchain.estimatefee", []interface{}{blocks})
	result := <-resultChan
	if result.Error != nil {
		return 0, result.Error
	}

	var fee float64
	if err := json.Unmarshal(result.Result, &fee); err != nil {
		log.ErrorWithContext(ctx, "解析手续费失败:", err)
		return 0, fmt.Errorf("解析手续费失败: %w", err)
	}

	return fee, nil
}

// ServerVersion 获取服务器版本信息
func ServerVersion(ctx context.Context) ([]interface{}, error) {
	resultChan := CallMethodAsync(ctx, "server.version", []interface{}{"electrumx-client", "1.4"})
	result := <-resultChan
	if result.Error != nil {
		return nil, result.Error
	}

	var version []interface{}
	if err := json.Unmarshal(result.Result, &version); err != nil {
		log.ErrorWithContext(ctx, "解析服务器版本信息失败:", err)
		return nil, fmt.Errorf("解析服务器版本信息失败: %w", err)
	}

	return version, nil
}

// ServerPeers 获取服务器的对等节点信息
func ServerPeers(ctx context.Context) ([]interface{}, error) {
	resultChan := CallMethodAsync(ctx, "server.peers.subscribe", []interface{}{})
	result := <-resultChan
	if result.Error != nil {
		return nil, result.Error
	}

	var peers []interface{}
	if err := json.Unmarshal(result.Result, &peers); err != nil {
		log.ErrorWithContext(ctx, "解析对等节点信息失败:", err)
		return nil, fmt.Errorf("解析对等节点信息失败: %w", err)
	}

	return peers, nil
}

// ServerFeatures 获取服务器特性
func ServerFeatures(ctx context.Context) (map[string]interface{}, error) {
	resultChan := CallMethodAsync(ctx, "server.features", []interface{}{})
	result := <-resultChan
	if result.Error != nil {
		return nil, result.Error
	}

	var features map[string]interface{}
	if err := json.Unmarshal(result.Result, &features); err != nil {
		log.ErrorWithContext(ctx, "解析服务器特性失败:", err)
		return nil, fmt.Errorf("解析服务器特性失败: %w", err)
	}

	return features, nil
}

// GetListUnspent 获取指定脚本哈希的未花费交易输出
func GetListUnspent(ctx context.Context, scriptHash string) (electrumx.UtxoResponse, error) {
	// 参数校验
	if len(scriptHash) == 0 {
		log.ErrorWithContext(ctx, "调用GetListUnspent失败: scriptHash不能为空")
		return nil, fmt.Errorf("scriptHash不能为空")
	}

	// 记录开始调用日志
	log.InfoWithContext(ctx, "开始获取脚本哈希的UTXO:", scriptHash)

	// 调用RPC方法（改为异步）
	resultChan := CallMethodAsync(ctx, "blockchain.scripthash.listunspent", []interface{}{scriptHash})
	result := <-resultChan
	if result.Error != nil {
		log.ErrorWithContext(ctx, "获取UTXO失败:", result.Error)
		return nil, fmt.Errorf("获取UTXO失败: %w", result.Error)
	}

	// 解析响应
	var utxos electrumx.UtxoResponse
	if err := json.Unmarshal(result.Result, &utxos); err != nil {
		log.ErrorWithContext(ctx, "解析UTXO响应失败:", err)
		return nil, fmt.Errorf("解析UTXO响应失败: %w", err)
	}

	log.InfoWithContext(ctx, "成功获取UTXO, 数量:", len(utxos))
	return utxos, nil
}

// GetScriptHashBalance 获取指定脚本哈希的余额
func GetScriptHashBalance(ctx context.Context, scriptHash string) (*electrumx.BalanceResponse, error) {
	// 参数校验
	if len(scriptHash) == 0 {
		log.ErrorWithContext(ctx, "调用GetScriptHashBalance失败: scriptHash不能为空")
		return nil, fmt.Errorf("scriptHash不能为空")
	}

	// 记录开始调用日志
	log.InfoWithContext(ctx, "开始获取脚本哈希余额:", scriptHash)

	// 调用RPC方法（改为异步）
	resultChan := CallMethodAsync(ctx, "blockchain.scripthash.get_balance", []interface{}{scriptHash})
	result := <-resultChan
	if result.Error != nil {
		log.ErrorWithContext(ctx, "获取脚本哈希余额失败:", result.Error)
		return nil, fmt.Errorf("获取脚本哈希余额失败: %w", result.Error)
	}

	// 解析响应
	var balance electrumx.BalanceResponse
	if err := json.Unmarshal(result.Result, &balance); err != nil {
		log.ErrorWithContext(ctx, "解析脚本哈希余额失败:", err, "原始数据:", string(result.Result))
		return nil, fmt.Errorf("解析脚本哈希余额失败: %w", err)
	}

	log.InfoWithContext(ctx, "成功获取脚本哈希余额: 已确认=", balance.Confirmed, "未确认=", balance.Unconfirmed)
	return &balance, nil
}

// GetScriptHashFrozenBalance 获取指定脚本哈希的冻结余额
func GetScriptHashFrozenBalance(ctx context.Context, scriptHash string) (*electrumx.FrozenBalanceResponse, error) {
	// 参数校验
	if len(scriptHash) == 0 {
		log.ErrorWithContext(ctx, "调用GetScriptHashFrozenBalance失败: scriptHash不能为空")
		return nil, fmt.Errorf("scriptHash不能为空")
	}

	// 记录开始调用日志
	log.InfoWithContext(ctx, "开始获取脚本哈希冻结余额:", scriptHash)

	// 调用RPC方法（改为异步）
	resultChan := CallMethodAsync(ctx, "blockchain.scripthash.get_frozen_balance", []interface{}{scriptHash})
	result := <-resultChan
	if result.Error != nil {
		log.ErrorWithContext(ctx, "获取脚本哈希冻结余额失败:", result.Error)
		return nil, fmt.Errorf("获取脚本哈希冻结余额失败: %w", result.Error)
	}

	// 解析响应
	var frozenBalance electrumx.FrozenBalanceResponse
	if err := json.Unmarshal(result.Result, &frozenBalance); err != nil {
		log.ErrorWithContext(ctx, "解析脚本哈希冻结余额失败:", err, "原始数据:", string(result.Result))
		return nil, fmt.Errorf("解析脚本哈希冻结余额失败: %w", err)
	}

	log.InfoWithContext(ctx, "成功获取脚本哈希冻结余额: 冻结=", frozenBalance.Frozen)
	return &frozenBalance, nil
}

// CallElectrumXRPC 通用的ElectrumX RPC调用函数，支持上下文控制
func CallElectrumXRPC(ctx context.Context, method string, params []interface{}) (json.RawMessage, error) {
	// 记录开始调用日志
	log.InfoWithContext(ctx, "开始调用ElectrumX RPC",
		"method:", method,
		"params:", params)

	// 使用异步方式调用
	resultChan := CallMethodAsync(ctx, method, params)

	// 等待结果
	select {
	case <-ctx.Done():
		// 上下文已取消（超时或其他原因）
		log.WarnWithContext(ctx, "ElectrumX RPC调用已取消",
			"method:", method,
			"错误:", ctx.Err())
		return nil, ctx.Err()
	case result := <-resultChan:
		// 收到结果
		if result.Error != nil {
			log.ErrorWithContext(ctx, "ElectrumX RPC调用失败",
				"method:", method,
				"错误:", result.Error)
			return nil, result.Error
		}

		// 记录成功调用
		log.InfoWithContext(ctx, "ElectrumX RPC调用成功",
			"method:", method,
			"responseSize:", len(result.Result))
		return result.Result, nil
	}
}

// GetBlockByHeight 获取指定高度的区块信息
func GetBlockByHeight(ctx context.Context, height int64) (*struct {
	Hash   string `json:"hash"`
	Height int64  `json:"height"`
	Time   int64  `json:"time"`
}, error) {
	// 参数校验
	if height < 0 {
		log.ErrorWithContext(ctx, "调用GetBlockByHeight失败: 区块高度不能为负数")
		return nil, fmt.Errorf("区块高度不能为负数")
	}

	log.InfoWithContext(ctx, "开始获取区块信息", "height", height)

	// 调用RPC方法获取区块头（改为异步）
	resultChan := CallMethodAsync(ctx, "blockchain.block.header", []interface{}{height})
	result := <-resultChan
	if result.Error != nil {
		log.ErrorWithContext(ctx, "获取区块头信息失败", "height", height, "error", result.Error)
		return nil, fmt.Errorf("获取区块头信息失败: %w", result.Error)
	}

	// 解析区块头信息
	var headerHex string
	if err := json.Unmarshal(result.Result, &headerHex); err != nil {
		log.ErrorWithContext(ctx, "解析区块头信息失败", "height", height, "error", err)
		return nil, fmt.Errorf("解析区块头信息失败: %w", err)
	}

	// 解析区块头十六进制数据
	// 这里简化处理，实际项目中需要根据比特币区块头格式进行正确解析
	// 区块头格式: Version(4) + PrevBlock(32) + MerkleRoot(32) + Time(4) + Bits(4) + Nonce(4)
	if len(headerHex) < 160 { // 原始区块头是80字节，十六进制表示为160个字符
		log.ErrorWithContext(ctx, "区块头数据长度不足", "height", height, "headerHex", headerHex)
		return nil, fmt.Errorf("区块头数据长度不足")
	}

	// 解析时间戳(第68-76个字符，对应时间戳字段，需要转换字节序)
	timeHex := headerHex[68:76]
	var timeBytes []byte
	for i := len(timeHex) - 2; i >= 0; i -= 2 {
		b, err := strconv.ParseUint(timeHex[i:i+2], 16, 8)
		if err != nil {
			log.ErrorWithContext(ctx, "解析时间戳失败", "height", height, "timeHex", timeHex, "error", err)
			return nil, fmt.Errorf("解析时间戳失败: %w", err)
		}
		timeBytes = append(timeBytes, byte(b))
	}
	timestamp := int64(binary.LittleEndian.Uint32(timeBytes))

	// 获取区块哈希（改为异步）
	hashResultChan := CallMethodAsync(ctx, "blockchain.block.header", []interface{}{height, 1})
	hashResult := <-hashResultChan
	if hashResult.Error != nil {
		log.ErrorWithContext(ctx, "获取区块哈希失败", "height", height, "error", hashResult.Error)
		return nil, fmt.Errorf("获取区块哈希失败: %w", hashResult.Error)
	}

	var hashObj struct {
		Hash string `json:"hash"`
	}
	if err := json.Unmarshal(hashResult.Result, &hashObj); err != nil {
		log.ErrorWithContext(ctx, "解析区块哈希失败", "height", height, "error", err)
		return nil, fmt.Errorf("解析区块哈希失败: %w", err)
	}

	// 构建返回结果
	blockInfo := &struct {
		Hash   string `json:"hash"`
		Height int64  `json:"height"`
		Time   int64  `json:"time"`
	}{
		Hash:   hashObj.Hash,
		Height: height,
		Time:   timestamp,
	}

	log.InfoWithContext(ctx, "成功获取区块信息", "height", height, "hash", blockInfo.Hash, "time", blockInfo.Time)
	return blockInfo, nil
}

// AddressToScriptHash 将比特币地址转换为脚本哈希
func AddressToScriptHash(ctx context.Context, address string) (string, error) {
	// 调用工具函数进行转换
	scriptHash, err := utility.AddressToScriptHash(address)
	if err != nil {
		log.ErrorWithContext(ctx, "地址转换为脚本哈希失败:", err)
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
	log.InfoWithContext(ctx, "开始获取脚本哈希的UTXO", "scriptHash:", scriptHash)

	// 参数校验
	if scriptHash == "" {
		log.ErrorWithContext(ctx, "脚本哈希不能为空")
		return nil, ErrEmptyScriptHash
	}

	// 调用ElectrumX RPC获取UTXO列表
	utxos, err := GetListUnspent(ctx, scriptHash)
	if err != nil {
		log.ErrorWithContext(ctx, "获取UTXO失败",
			"scriptHash:", scriptHash,
			"错误:", err)
		return nil, err
	}

	log.InfoWithContext(ctx, "成功获取UTXO",
		"scriptHash:", scriptHash,
		"count:", len(utxos))
	return utxos, nil
}

// ConvertAddressToScript 将WIF地址转换为脚本哈希
func ConvertAddressToScript(ctx context.Context, address string) (string, error) {
	// 调用AddressToScriptHash函数进行转换
	scriptHash, err := AddressToScriptHash(ctx, address)
	if err != nil {
		log.ErrorWithContext(ctx, "地址转换为脚本哈希失败:", err)
		return "", err
	}

	return scriptHash, nil
}

// GetScriptHashHistory 使用上下文获取指定脚本哈希的交易历史
func GetScriptHashHistory(ctx context.Context, scriptHash string) (electrumx.ElectrumXHistoryResponse, error) {
	// 参数校验
	if scriptHash == "" {
		log.ErrorWithContext(ctx, "脚本哈希不能为空")
		return nil, ErrEmptyScriptHash
	}

	// 记录调用开始
	log.InfoWithContext(ctx, "开始获取脚本哈希历史",
		"scriptHash:", scriptHash)

	// 通过通用调用函数执行RPC请求
	result, err := CallElectrumXRPC(ctx, "blockchain.scripthash.get_history", []interface{}{scriptHash})
	if err != nil {
		log.ErrorWithContext(ctx, "获取脚本哈希历史失败",
			"scriptHash:", scriptHash,
			"错误:", err)
		return nil, fmt.Errorf("获取脚本哈希历史失败: %w", err)
	}

	// 解析响应数据
	var history electrumx.ElectrumXHistoryResponse
	if err := json.Unmarshal(result, &history); err != nil {
		log.ErrorWithContext(ctx, "解析脚本哈希历史失败",
			"scriptHash:", scriptHash,
			"错误:", err)
		return nil, fmt.Errorf("解析脚本哈希历史失败: %w", err)
	}

	// 记录成功获取
	log.InfoWithContext(ctx, "成功获取脚本哈希历史",
		"scriptHash:", scriptHash,
		"count:", len(history))

	return history, nil
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
		"address:", address)

	// 将地址转换为脚本哈希
	scriptHash, err := AddressToScriptHash(ctx, address)
	if err != nil {
		log.ErrorWithContext(ctx, "地址转换为脚本哈希失败",
			"address:", address,
			"错误:", err)
		return nil, fmt.Errorf("地址转换为脚本哈希失败: %w", err)
	}

	// 获取脚本哈希余额
	balance, err := GetScriptHashBalance(ctx, scriptHash)
	if err != nil {
		log.ErrorWithContext(ctx, "获取地址余额失败",
			"address:", address,
			"scriptHash:", scriptHash,
			"错误:", err)
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
		"address:", address,
		"balance:", totalBalance,
		"confirmed:", balance.Confirmed,
		"unconfirmed:", balance.Unconfirmed)

	return response, nil
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
		"address:", address)

	// 将地址转换为脚本哈希
	scriptHash, err := AddressToScriptHash(ctx, address)
	if err != nil {
		log.ErrorWithContext(ctx, "地址转换为脚本哈希失败",
			"address:", address,
			"错误:", err)
		return nil, fmt.Errorf("地址转换为脚本哈希失败: %w", err)
	}

	// 获取脚本哈希冻结余额
	frozenBalance, err := GetScriptHashFrozenBalance(ctx, scriptHash)
	if err != nil {
		log.ErrorWithContext(ctx, "获取地址冻结余额失败",
			"address:", address,
			"scriptHash:", scriptHash,
			"错误:", err)
		return nil, fmt.Errorf("获取地址冻结余额失败: %w", err)
	}

	log.InfoWithContext(ctx, "成功获取地址冻结余额",
		"address:", address,
		"frozen:", frozenBalance.Frozen)

	return frozenBalance, nil
}

// GetScriptHashUnspent 使用上下文获取指定脚本哈希的未花费交易输出
func GetScriptHashUnspent(ctx context.Context, scriptHash string) (electrumx.UtxoResponse, error) {
	// 参数校验
	if scriptHash == "" {
		log.ErrorWithContext(ctx, "脚本哈希不能为空")
		return nil, ErrEmptyScriptHash
	}

	// 记录调用开始
	log.InfoWithContext(ctx, "开始获取脚本哈希的UTXO",
		"scriptHash:", scriptHash)

	// 通过通用调用函数执行RPC请求
	result, err := CallElectrumXRPC(ctx, "blockchain.scripthash.listunspent", []interface{}{scriptHash})
	if err != nil {
		log.ErrorWithContext(ctx, "获取脚本哈希UTXO失败",
			"scriptHash:", scriptHash,
			"错误:", err)
		return nil, fmt.Errorf("获取脚本哈希UTXO失败: %w", err)
	}

	// 解析响应数据
	var utxos electrumx.UtxoResponse
	if err := json.Unmarshal(result, &utxos); err != nil {
		log.ErrorWithContext(ctx, "解析脚本哈希UTXO失败",
			"scriptHash:", scriptHash,
			"错误:", err)
		return nil, fmt.Errorf("解析脚本哈希UTXO失败: %w", err)
	}

	// 记录成功获取
	log.InfoWithContext(ctx, "成功获取脚本哈希UTXO",
		"scriptHash:", scriptHash,
		"count:", len(utxos))

	return utxos, nil
}
