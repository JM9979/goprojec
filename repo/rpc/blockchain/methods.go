package blockchain

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"ginproject/entity/blockchain"
	"ginproject/middleware/log"
)

// 节点RPC方法名常量
const (
	RpcMethodGetBlockByHeight  = "getblockbyheight"
	RpcMethodGetBlock          = "getblock"
	RpcMethodGetBlockHash      = "getblockhash"
	RpcMethodGetBlockHeader    = "getblockheader"
	RpcMethodGetInfo           = "getinfo"
	RpcMethodGetBlockchainInfo = "getblockchaininfo"
)

// BlockInfo 表示区块信息
type BlockInfo struct {
	Hash          string   `json:"hash"`
	Confirmations int      `json:"confirmations"`
	Size          int      `json:"size"`
	Height        int      `json:"height"`
	Version       int      `json:"version"`
	MerkleRoot    string   `json:"merkleroot"`
	Tx            []string `json:"tx"`
	Time          int64    `json:"time"`
	Nonce         int      `json:"nonce"`
	Bits          string   `json:"bits"`
	Difficulty    float64  `json:"difficulty"`
	PreviousHash  string   `json:"previousblockhash,omitempty"`
	NextHash      string   `json:"nextblockhash,omitempty"`
}

// GetBlockByHeightStructured 根据区块高度获取结构化区块信息（异步）
func GetBlockByHeightStructured(ctx context.Context, height int) <-chan AsyncResult {
	resultChan := make(chan AsyncResult, 1)

	go func() {
		defer close(resultChan)

		// 参数验证
		if height < 0 {
			log.ErrorWithContext(ctx, "获取区块信息失败：区块高度不能为负数", "height:", height)
			resultChan <- AsyncResult{
				Result: nil,
				Error:  fmt.Errorf("区块高度不能为负数"),
			}
			return
		}

		// 记录开始调用日志
		log.InfoWithContext(ctx, "开始获取区块信息", "height:", height)

		// 调用区块链节点RPC
		result, err := CallRPC("getblockbyheight", []interface{}{height, true}, false)
		if err != nil {
			log.ErrorWithContext(ctx, "获取区块信息失败", "height:", height, "错误:", err)
			resultChan <- AsyncResult{
				Result: nil,
				Error:  fmt.Errorf("获取区块信息失败: %w", err),
			}
			return
		}

		// 将返回结果转换为BlockInfo结构
		var blockInfo BlockInfo
		resultBytes, err := json.Marshal(result)
		if err != nil {
			log.ErrorWithContext(ctx, "序列化区块结果失败", "错误:", err)
			resultChan <- AsyncResult{
				Result: nil,
				Error:  fmt.Errorf("解析RPC响应失败: %w", err),
			}
			return
		}

		if err := json.Unmarshal(resultBytes, &blockInfo); err != nil {
			log.ErrorWithContext(ctx, "解析区块数据失败", "height:", height, "错误:", err)
			resultChan <- AsyncResult{
				Result: nil,
				Error:  fmt.Errorf("解析区块数据失败: %w", err),
			}
			return
		}

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

		log.InfoWithContext(ctx, "成功获取区块信息",
			"height:", height,
			"hash:", blockInfo.Hash,
			"time:", blockInfo.Time)
		resultChan <- AsyncResult{
			Result: &blockInfo,
			Error:  nil,
		}
	}()

	return resultChan
}

// FetchBlockByHeight 根据区块高度获取区块详情(原始接口数据)（异步）
func FetchBlockByHeight(ctx context.Context, height int64) <-chan AsyncResult {
	resultChan := make(chan AsyncResult, 1)

	go func() {
		defer close(resultChan)

		traceId := ctx.Value("trace_id")
		log.Info("[RPC请求] 通过高度获取区块", "trace_id", traceId, "height", height)

		response, err := CallRPC(RpcMethodGetBlockByHeight, []interface{}{height}, false)
		if err != nil {
			log.Error("[RPC请求失败] 通过高度获取区块", "trace_id", traceId, "height", height, "error", err)
			resultChan <- AsyncResult{
				Result: nil,
				Error:  err,
			}
			return
		}

		responseMap, ok := response.(map[string]interface{})
		if !ok {
			err := fmt.Errorf("响应格式错误")
			log.Error("[RPC响应格式错误] 通过高度获取区块", "trace_id", traceId, "height", height)
			resultChan <- AsyncResult{
				Result: nil,
				Error:  err,
			}
			return
		}

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

		log.Info("[RPC请求成功] 通过高度获取区块", "trace_id", traceId, "height", height)
		resultChan <- AsyncResult{
			Result: responseMap,
			Error:  nil,
		}
	}()

	return resultChan
}

// FetchBlockByHash 根据区块哈希获取区块详情(原始接口数据)（异步）
func FetchBlockByHash(ctx context.Context, hash string) <-chan AsyncResult {
	resultChan := make(chan AsyncResult, 1)

	go func() {
		defer close(resultChan)

		traceId := ctx.Value("trace_id")
		log.Info("[RPC请求] 通过哈希获取区块", "trace_id", traceId, "hash", hash)

		response, err := CallRPC(RpcMethodGetBlock, []interface{}{hash}, false)
		if err != nil {
			log.Error("[RPC请求失败] 通过哈希获取区块", "trace_id", traceId, "hash", hash, "error", err)
			resultChan <- AsyncResult{
				Result: nil,
				Error:  err,
			}
			return
		}

		responseMap, ok := response.(map[string]interface{})
		if !ok {
			err := fmt.Errorf("响应格式错误")
			log.Error("[RPC响应格式错误] 通过哈希获取区块", "trace_id", traceId, "hash", hash)
			resultChan <- AsyncResult{
				Result: nil,
				Error:  err,
			}
			return
		}

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

		log.Info("[RPC请求成功] 通过哈希获取区块", "trace_id", traceId, "hash", hash)
		resultChan <- AsyncResult{
			Result: responseMap,
			Error:  nil,
		}
	}()

	return resultChan
}

// FetchBlockHeaderByHeight 根据区块高度获取区块头信息（异步）
func FetchBlockHeaderByHeight(ctx context.Context, height int64) <-chan AsyncResult {
	resultChan := make(chan AsyncResult, 1)

	go func() {
		defer close(resultChan)

		traceId := ctx.Value("trace_id")
		log.Info("[RPC请求] 通过高度获取区块头", "trace_id", traceId, "height", height)

		// 先获取区块哈希
		hashResponse, err := CallRPC(RpcMethodGetBlockHash, []interface{}{height}, false)
		if err != nil {
			log.Error("[RPC请求失败] 获取区块哈希", "trace_id", traceId, "height", height, "error", err)
			resultChan <- AsyncResult{
				Result: nil,
				Error:  err,
			}
			return
		}

		hash, ok := hashResponse.(string)
		if !ok {
			err := fmt.Errorf("响应格式错误")
			log.Error("[RPC响应格式错误] 获取区块哈希", "trace_id", traceId, "height", height)
			resultChan <- AsyncResult{
				Result: nil,
				Error:  err,
			}
			return
		}

		// 通过哈希获取区块头
		response, err := CallRPC(RpcMethodGetBlockHeader, []interface{}{hash}, false)
		if err != nil {
			log.Error("[RPC请求失败] 通过哈希获取区块头", "trace_id", traceId, "hash", hash, "error", err)
			resultChan <- AsyncResult{
				Result: nil,
				Error:  err,
			}
			return
		}

		responseMap, ok := response.(map[string]interface{})
		if !ok {
			err := fmt.Errorf("响应格式错误")
			log.Error("[RPC响应格式错误] 通过哈希获取区块头", "trace_id", traceId, "height", height)
			resultChan <- AsyncResult{
				Result: nil,
				Error:  err,
			}
			return
		}

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

		log.Info("[RPC请求成功] 通过高度获取区块头", "trace_id", traceId, "height", height)
		resultChan <- AsyncResult{
			Result: responseMap,
			Error:  nil,
		}
	}()

	return resultChan
}

// FetchBlockHeaderByHash 根据区块哈希获取区块头信息（异步）
func FetchBlockHeaderByHash(ctx context.Context, hash string) <-chan AsyncResult {
	resultChan := make(chan AsyncResult, 1)

	go func() {
		defer close(resultChan)

		traceId := ctx.Value("trace_id")
		log.Info("[RPC请求] 通过哈希获取区块头", "trace_id", traceId, "hash", hash)

		response, err := CallRPC(RpcMethodGetBlockHeader, []interface{}{hash}, false)
		if err != nil {
			log.Error("[RPC请求失败] 通过哈希获取区块头", "trace_id", traceId, "hash", hash, "error", err)
			resultChan <- AsyncResult{
				Result: nil,
				Error:  err,
			}
			return
		}

		responseMap, ok := response.(map[string]interface{})
		if !ok {
			err := fmt.Errorf("响应格式错误")
			log.Error("[RPC响应格式错误] 通过哈希获取区块头", "trace_id", traceId, "hash", hash)
			resultChan <- AsyncResult{
				Result: nil,
				Error:  err,
			}
			return
		}

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

		log.Info("[RPC请求成功] 通过哈希获取区块头", "trace_id", traceId, "hash", hash)
		resultChan <- AsyncResult{
			Result: responseMap,
			Error:  nil,
		}
	}()

	return resultChan
}

// FetchNearby10Headers 获取最近的10个区块头信息（异步）
func FetchNearby10Headers(ctx context.Context) <-chan AsyncResult {
	resultChan := make(chan AsyncResult, 1)

	go func() {
		defer close(resultChan)

		traceId := ctx.Value("trace_id")
		log.Info("[RPC请求] 获取最近10个区块头", "trace_id", traceId)

		// 获取当前区块高度
		infoResponse, err := CallRPC(RpcMethodGetInfo, []interface{}{}, false)
		if err != nil {
			log.Error("[RPC请求失败] 获取区块链信息", "trace_id", traceId, "error", err)
			resultChan <- AsyncResult{
				Result: nil,
				Error:  err,
			}
			return
		}

		info, ok := infoResponse.(map[string]interface{})
		if !ok {
			err := fmt.Errorf("响应格式错误")
			log.Error("[RPC响应格式错误] 获取区块链信息", "trace_id", traceId)
			resultChan <- AsyncResult{
				Result: nil,
				Error:  err,
			}
			return
		}

		height, ok := info["blocks"].(float64)
		if !ok {
			err := fmt.Errorf("响应格式错误")
			log.Error("[RPC响应格式错误] 获取区块高度", "trace_id", traceId)
			resultChan <- AsyncResult{
				Result: nil,
				Error:  err,
			}
			return
		}

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

		// 获取最近10个区块头
		var response []map[string]interface{}
		for i := 0; i < 10; i++ {
			blockHeight := int64(height) - int64(i)
			if blockHeight < 0 {
				break
			}

			// 使用同步方式调用，因为我们需要按顺序获取
			headerChan := FetchBlockHeaderByHeight(ctx, blockHeight)
			headerResult := <-headerChan

			if headerResult.Error != nil {
				log.Error("[RPC请求失败] 获取区块头", "trace_id", traceId, "height", blockHeight, "error", headerResult.Error)
				continue
			}

			headerMap, ok := headerResult.Result.(map[string]interface{})
			if ok {
				response = append(response, headerMap)
			}

			// 再次检查上下文是否已取消
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
		}

		log.Info("[RPC请求成功] 获取最近10个区块头", "trace_id", traceId, "count", len(response))
		resultChan <- AsyncResult{
			Result: response,
			Error:  nil,
		}
	}()

	return resultChan
}

// GetRawTransaction 获取交易原始信息（异步）
// 直接返回RPC调用结果，不进行结构体转换
func GetRawTransaction(ctx context.Context, txid string, verbose bool) <-chan AsyncResult {
	resultChan := make(chan AsyncResult, 1)

	go func() {
		defer close(resultChan)

		verboseInt := 0
		if verbose {
			verboseInt = 1
		}

		log.InfoWithContextf(ctx, "开始获取交易原始信息: txid=%s, verbose=%v", txid, verbose)
		result, err := CallRPC("getrawtransaction", []interface{}{txid, verboseInt}, false)
		if err != nil {
			log.ErrorWithContextf(ctx, "获取交易原始信息失败: txid=%s, 错误=%v", txid, err)
			resultChan <- AsyncResult{
				Result: nil,
				Error:  fmt.Errorf("获取交易失败: %w", err),
			}
			return
		}

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

		log.InfoWithContextf(ctx, "成功获取交易原始信息: txid=%s", txid)
		resultChan <- AsyncResult{
			Result: result,
			Error:  nil,
		}
	}()

	return resultChan
}

// GetBlockByHeight 根据区块高度获取区块信息(简化版)（异步）
func GetBlockByHeight(ctx context.Context, height int64) <-chan AsyncResult {
	resultChan := make(chan AsyncResult, 1)

	go func() {
		defer close(resultChan)

		log.InfoWithContextf(ctx, "开始获取区块信息: height=%d", height)
		result, err := CallRPC("getblockbyheight", []interface{}{height}, false)
		if err != nil {
			log.ErrorWithContextf(ctx, "获取区块信息失败: height=%d, 错误=%v", height, err)
			resultChan <- AsyncResult{
				Result: nil,
				Error:  fmt.Errorf("获取区块失败: %w", err),
			}
			return
		}

		resultMap, ok := result.(map[string]interface{})
		if !ok {
			log.ErrorWithContextf(ctx, "区块信息格式不正确: height=%d", height)
			resultChan <- AsyncResult{
				Result: nil,
				Error:  fmt.Errorf("区块信息格式不正确"),
			}
			return
		}

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

		log.InfoWithContextf(ctx, "成功获取区块信息: height=%d", height)
		resultChan <- AsyncResult{
			Result: resultMap,
			Error:  nil,
		}
	}()

	return resultChan
}

// FetchChainInfo 获取区块链信息（异步）
func FetchChainInfo(ctx context.Context) <-chan AsyncResult {
	resultChan := make(chan AsyncResult, 1)

	go func() {
		defer close(resultChan)

		traceId := ctx.Value("trace_id")
		log.Info("[RPC请求] 获取区块链信息", "trace_id", traceId)

		response, err := CallRPC(RpcMethodGetBlockchainInfo, []interface{}{}, false)
		if err != nil {
			log.Error("[RPC请求失败] 获取区块链信息", "trace_id", traceId, "error", err)
			resultChan <- AsyncResult{
				Result: nil,
				Error:  err,
			}
			return
		}

		responseMap, ok := response.(map[string]interface{})
		if !ok {
			err := fmt.Errorf("响应格式错误")
			log.Error("[RPC响应格式错误] 获取区块链信息", "trace_id", traceId)
			resultChan <- AsyncResult{
				Result: nil,
				Error:  err,
			}
			return
		}

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

		log.Info("[RPC请求成功] 获取区块链信息", "trace_id", traceId)
		resultChan <- AsyncResult{
			Result: responseMap,
			Error:  nil,
		}
	}()

	return resultChan
}

// DecodeTxHash 根据交易哈希获取交易详情（异步）
// 此方法通过区块链节点RPC接口查询指定交易哈希的详细信息
// 返回的交易信息包括输入输出、脚本、金额等详细数据
// 参数:
//   - ctx: 上下文，用于控制请求的生命周期
//   - txid: 交易哈希ID
//
// 返回:
//   - <-chan AsyncResult: 包含交易详情的异步结果通道
func DecodeTxHash(ctx context.Context, txid string) <-chan AsyncResult {
	resultChan := make(chan AsyncResult, 1)

	go func() {
		defer close(resultChan)

		// 参数验证
		if txid == "" {
			log.ErrorWithContextf(ctx, "解析交易失败: 交易ID不能为空")
			resultChan <- AsyncResult{
				Result: nil,
				Error:  errors.New("交易ID不能为空"),
			}
			return
		}

		// 记录开始调用日志
		log.InfoWithContextf(ctx, "开始查询交易: %s", txid)

		// 调用区块链节点RPC
		result, err := CallRPC("getrawtransaction", []interface{}{txid, 1}, false)
		if err != nil {
			log.ErrorWithContextf(ctx, "查询交易失败: %s, 错误: %v", txid, err)
			resultChan <- AsyncResult{
				Result: nil,
				Error:  fmt.Errorf("解析交易失败: %w", err),
			}
			return
		}

		// 将返回结果转换为TransactionResponse结构
		var tx blockchain.TransactionResponse
		resultBytes, err := json.Marshal(result)
		if err != nil {
			log.ErrorWithContextf(ctx, "序列化交易结果失败: %v", err)
			resultChan <- AsyncResult{
				Result: nil,
				Error:  fmt.Errorf("解析RPC响应失败: %w", err),
			}
			return
		}

		if err := json.Unmarshal(resultBytes, &tx); err != nil {
			log.ErrorWithContextf(ctx, "解析交易数据失败: %s, 错误: %v", txid, err)
			resultChan <- AsyncResult{
				Result: nil,
				Error:  fmt.Errorf("解析交易数据失败: %w", err),
			}
			return
		}

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

		log.InfoWithContextf(ctx, "成功查询交易: %s, 确认数: %d", txid, tx.Confirmations)
		resultChan <- AsyncResult{
			Result: &tx,
			Error:  nil,
		}
	}()

	return resultChan
}

// DecodeTx 根据交易哈希获取交易详情（异步）
// 此方法通过区块链节点RPC接口查询指定交易哈希的详细信息
// 返回的交易信息包括输入输出、脚本、金额等详细数据
func DecodeTx(ctx context.Context, txid string) <-chan AsyncResult {
	resultChan := make(chan AsyncResult, 1)

	go func() {
		defer close(resultChan)

		// 参数验证
		if txid == "" {
			log.ErrorWithContextf(ctx, "解析交易失败: 交易ID不能为空")
			resultChan <- AsyncResult{
				Result: nil,
				Error:  fmt.Errorf("交易ID不能为空"),
			}
			return
		}

		// 记录开始调用日志
		log.InfoWithContextf(ctx, "开始查询交易: %s", txid)

		// 调用区块链节点RPC
		result, err := CallRPC("getrawtransaction", []interface{}{txid, 1}, false)
		if err != nil {
			log.ErrorWithContextf(ctx, "查询交易失败: %s, 错误: %v", txid, err)
			resultChan <- AsyncResult{
				Result: nil,
				Error:  fmt.Errorf("解析交易失败: %w", err),
			}
			return
		}

		// 将返回结果转换为TransactionResponse结构
		var tx blockchain.TransactionResponse
		resultBytes, err := json.Marshal(result)
		if err != nil {
			log.ErrorWithContextf(ctx, "序列化交易结果失败: %v", err)
			resultChan <- AsyncResult{
				Result: nil,
				Error:  fmt.Errorf("解析RPC响应失败: %w", err),
			}
			return
		}

		if err := json.Unmarshal(resultBytes, &tx); err != nil {
			log.ErrorWithContextf(ctx, "解析交易数据失败: %s, 错误: %v", txid, err)
			resultChan <- AsyncResult{
				Result: nil,
				Error:  fmt.Errorf("解析交易数据失败: %w", err),
			}
			return
		}

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

		log.InfoWithContextf(ctx, "成功查询交易: %s, 确认数: %d", txid, tx.Confirmations)
		resultChan <- AsyncResult{
			Result: &tx,
			Error:  nil,
		}
	}()

	return resultChan
}

// DecodeRawTransaction 根据交易哈希获取原始交易详情（异步）
// 此方法直接返回节点返回的原始结果，不进行结构体转换
// 等同于Python中的decode_tx_hash函数
func DecodeRawTransaction(ctx context.Context, txid string) <-chan AsyncResult {
	resultChan := make(chan AsyncResult, 1)

	go func() {
		defer close(resultChan)

		// 参数验证
		if txid == "" {
			log.ErrorWithContextf(ctx, "解析交易失败: 交易ID不能为空")
			resultChan <- AsyncResult{
				Result: nil,
				Error:  fmt.Errorf("交易ID不能为空"),
			}
			return
		}

		// 记录开始调用日志
		log.InfoWithContextf(ctx, "开始查询原始交易: %s", txid)

		// 调用区块链节点RPC
		result, err := CallRPC("getrawtransaction", []interface{}{txid, 1}, false)
		if err != nil {
			log.ErrorWithContextf(ctx, "查询原始交易失败: %s, 错误: %v", txid, err)
			resultChan <- AsyncResult{
				Result: nil,
				Error:  fmt.Errorf("解析交易失败: %w", err),
			}
			return
		}

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

		log.InfoWithContextf(ctx, "成功查询原始交易: %s", txid)
		resultChan <- AsyncResult{
			Result: result,
			Error:  nil,
		}
	}()

	return resultChan
}
