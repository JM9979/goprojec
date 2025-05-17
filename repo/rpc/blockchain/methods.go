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
	RpcMethodGetBlockByHeight     = "getblockbyheight"
	RpcMethodGetBlock             = "getblock"
	RpcMethodGetBlockHash         = "getblockhash"
	RpcMethodGetBlockHeader       = "getblockheader"
	RpcMethodGetInfo              = "getinfo"
	RpcMethodGetBlockchainInfo    = "getblockchaininfo"
	RpcMethodGetRawMempool        = "getrawmempool"
	RpcMethodGetRawTransaction    = "getrawtransaction"
	RpcMethodDecodeRawTransaction = "decoderawtransaction"
	RpcMethodSendRawTransaction   = "sendrawtransaction"
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

		// 调用区块链节点RPC（使用异步方式）
		asyncChan := CallRPCAsync(ctx, "getblockbyheight", []interface{}{height, true}, false)
		asyncResult := <-asyncChan

		if asyncResult.Error != nil {
			log.ErrorWithContext(ctx, "获取区块信息失败", "height:", height, "错误:", asyncResult.Error)
			resultChan <- AsyncResult{
				Result: nil,
				Error:  fmt.Errorf("获取区块信息失败: %w", asyncResult.Error),
			}
			return
		}

		result := asyncResult.Result

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

		log.InfoWithContext(ctx, "通过高度获取区块", "height", height)

		// 使用异步RPC调用
		responseChan := CallRPCAsync(ctx, RpcMethodGetBlockByHeight, []interface{}{height}, false)
		asyncResult := <-responseChan

		if asyncResult.Error != nil {
			log.ErrorWithContext(ctx, "通过高度获取区块失败", "height", height, "error", asyncResult.Error)
			resultChan <- AsyncResult{
				Result: nil,
				Error:  asyncResult.Error,
			}
			return
		}

		response := asyncResult.Result

		responseMap, ok := response.(map[string]interface{})
		if !ok {
			err := fmt.Errorf("响应格式错误")
			log.ErrorWithContext(ctx, "通过高度获取区块响应格式错误", "height", height)
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

		log.InfoWithContext(ctx, "通过高度获取区块成功", "height", height)
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

		log.InfoWithContext(ctx, "通过哈希获取区块", "hash", hash)

		// 使用异步RPC调用
		responseChan := CallRPCAsync(ctx, RpcMethodGetBlock, []interface{}{hash}, false)
		asyncResult := <-responseChan

		if asyncResult.Error != nil {
			log.ErrorWithContext(ctx, "通过哈希获取区块失败", "hash", hash, "error", asyncResult.Error)
			resultChan <- AsyncResult{
				Result: nil,
				Error:  asyncResult.Error,
			}
			return
		}

		response := asyncResult.Result

		responseMap, ok := response.(map[string]interface{})
		if !ok {
			err := fmt.Errorf("响应格式错误")
			log.ErrorWithContext(ctx, "通过哈希获取区块响应格式错误", "hash", hash)
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

		log.InfoWithContext(ctx, "通过哈希获取区块成功", "hash", hash)
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

		log.InfoWithContext(ctx, "通过高度获取区块头", "height", height)

		// 先获取区块哈希（使用异步方式）
		hashAsyncChan := CallRPCAsync(ctx, RpcMethodGetBlockHash, []interface{}{height}, false)
		hashAsyncResult := <-hashAsyncChan

		if hashAsyncResult.Error != nil {
			log.ErrorWithContext(ctx, "获取区块哈希失败", "height", height, "error", hashAsyncResult.Error)
			resultChan <- AsyncResult{
				Result: nil,
				Error:  hashAsyncResult.Error,
			}
			return
		}

		hash, ok := hashAsyncResult.Result.(string)
		if !ok {
			err := fmt.Errorf("响应格式错误")
			log.ErrorWithContext(ctx, "获取区块哈希响应格式错误", "height", height)
			resultChan <- AsyncResult{
				Result: nil,
				Error:  err,
			}
			return
		}

		// 通过哈希获取区块头（使用异步方式）
		responseAsyncChan := CallRPCAsync(ctx, RpcMethodGetBlockHeader, []interface{}{hash}, false)
		responseAsyncResult := <-responseAsyncChan

		if responseAsyncResult.Error != nil {
			log.ErrorWithContext(ctx, "通过哈希获取区块头失败", "hash", hash, "error", responseAsyncResult.Error)
			resultChan <- AsyncResult{
				Result: nil,
				Error:  responseAsyncResult.Error,
			}
			return
		}

		response := responseAsyncResult.Result

		responseMap, ok := response.(map[string]interface{})
		if !ok {
			err := fmt.Errorf("响应格式错误")
			log.ErrorWithContext(ctx, "通过哈希获取区块头响应格式错误", "height", height)
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

		log.InfoWithContext(ctx, "通过高度获取区块头成功", "height", height)
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

		log.InfoWithContext(ctx, "通过哈希获取区块头", "hash", hash)

		// 使用异步方式调用RPC
		asyncChan := CallRPCAsync(ctx, RpcMethodGetBlockHeader, []interface{}{hash}, false)
		asyncResult := <-asyncChan

		if asyncResult.Error != nil {
			log.ErrorWithContext(ctx, "通过哈希获取区块头失败", "hash", hash, "error", asyncResult.Error)
			resultChan <- AsyncResult{
				Result: nil,
				Error:  asyncResult.Error,
			}
			return
		}

		response := asyncResult.Result

		responseMap, ok := response.(map[string]interface{})
		if !ok {
			err := fmt.Errorf("响应格式错误")
			log.ErrorWithContext(ctx, "通过哈希获取区块头响应格式错误", "hash", hash)
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

		log.InfoWithContext(ctx, "通过哈希获取区块头成功", "hash", hash)
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

		log.InfoWithContext(ctx, "获取最近10个区块头")

		// 获取当前区块高度（使用异步方式）
		infoAsyncChan := CallRPCAsync(ctx, RpcMethodGetInfo, []interface{}{}, false)
		infoAsyncResult := <-infoAsyncChan

		if infoAsyncResult.Error != nil {
			log.ErrorWithContext(ctx, "获取区块链信息失败", "error", infoAsyncResult.Error)
			resultChan <- AsyncResult{
				Result: nil,
				Error:  infoAsyncResult.Error,
			}
			return
		}

		info, ok := infoAsyncResult.Result.(map[string]interface{})
		if !ok {
			err := fmt.Errorf("响应格式错误")
			log.ErrorWithContext(ctx, "获取区块链信息响应格式错误")
			resultChan <- AsyncResult{
				Result: nil,
				Error:  err,
			}
			return
		}

		height, ok := info["blocks"].(float64)
		if !ok {
			err := fmt.Errorf("响应格式错误")
			log.ErrorWithContext(ctx, "获取区块高度响应格式错误")
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
				log.ErrorWithContext(ctx, "获取区块头失败", "height", blockHeight, "error", headerResult.Error)
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

		log.InfoWithContext(ctx, "获取最近10个区块头成功", "count", len(response))
		resultChan <- AsyncResult{
			Result: response,
			Error:  nil,
		}
	}()

	return resultChan
}

// GetRawTransaction 获取交易原始数据
func GetRawTransaction(ctx context.Context, txid string, verbose bool) <-chan AsyncResult {
	resultChan := make(chan AsyncResult, 1)

	go func() {
		defer close(resultChan)

		// 参数验证
		if txid == "" {
			log.ErrorWithContext(ctx, "获取交易原始数据失败：交易ID不能为空")
			resultChan <- AsyncResult{
				Result: nil,
				Error:  fmt.Errorf("交易ID不能为空"),
			}
			return
		}

		// 记录开始调用日志
		log.InfoWithContext(ctx, "开始获取交易原始数据", "txid", txid, "verbose", verbose)

		// 准备参数
		params := []interface{}{txid}
		if verbose {
			params = append(params, 1)
		}

		// 调用区块链节点RPC（使用异步方式）
		asyncChan := CallRPCAsync(ctx, RpcMethodGetRawTransaction, params, false)
		asyncResult := <-asyncChan

		if asyncResult.Error != nil {
			log.ErrorWithContext(ctx, "获取交易原始数据失败", "txid", txid, "错误", asyncResult.Error)
			resultChan <- AsyncResult{
				Result: nil,
				Error:  fmt.Errorf("获取交易原始数据失败: %w", asyncResult.Error),
			}
			return
		}

		result := asyncResult.Result

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

		log.InfoWithContext(ctx, "成功获取交易原始数据", "txid", txid)
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

		// 使用异步RPC调用
		asyncChan := CallRPCAsync(ctx, "getblockbyheight", []interface{}{height}, false)
		asyncResult := <-asyncChan

		if asyncResult.Error != nil {
			log.ErrorWithContextf(ctx, "获取区块信息失败: height=%d, 错误=%v", height, asyncResult.Error)
			resultChan <- AsyncResult{
				Result: nil,
				Error:  fmt.Errorf("获取区块失败: %w", asyncResult.Error),
			}
			return
		}

		result := asyncResult.Result

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

		log.InfoWithContext(ctx, "获取区块链信息")

		// 使用异步方式调用RPC
		asyncChan := CallRPCAsync(ctx, RpcMethodGetBlockchainInfo, []interface{}{}, false)
		asyncResult := <-asyncChan

		if asyncResult.Error != nil {
			log.ErrorWithContext(ctx, "获取区块链信息失败", "error", asyncResult.Error)
			resultChan <- AsyncResult{
				Result: nil,
				Error:  asyncResult.Error,
			}
			return
		}

		response := asyncResult.Result

		responseMap, ok := response.(map[string]interface{})
		if !ok {
			err := fmt.Errorf("响应格式错误")
			log.ErrorWithContext(ctx, "获取区块链信息响应格式错误")
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

		log.InfoWithContext(ctx, "获取区块链信息成功")
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

		// 调用区块链节点RPC（使用异步方式）
		asyncChan := CallRPCAsync(ctx, "getrawtransaction", []interface{}{txid, 1}, false)
		asyncResult := <-asyncChan

		if asyncResult.Error != nil {
			log.ErrorWithContextf(ctx, "查询交易失败: %s, 错误: %v", txid, asyncResult.Error)
			resultChan <- AsyncResult{
				Result: nil,
				Error:  fmt.Errorf("解析交易失败: %w", asyncResult.Error),
			}
			return
		}

		result := asyncResult.Result

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

		// 调用区块链节点RPC（使用异步方式）
		asyncChan := CallRPCAsync(ctx, "getrawtransaction", []interface{}{txid, 1}, false)
		asyncResult := <-asyncChan

		if asyncResult.Error != nil {
			log.ErrorWithContextf(ctx, "查询交易失败: %s, 错误: %v", txid, asyncResult.Error)
			resultChan <- AsyncResult{
				Result: nil,
				Error:  fmt.Errorf("解析交易失败: %w", asyncResult.Error),
			}
			return
		}

		result := asyncResult.Result

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

// DecodeRawTransaction 解码原始交易
func DecodeRawTransaction(ctx context.Context, txHex string) <-chan AsyncResult {
	resultChan := make(chan AsyncResult, 1)

	go func() {
		defer close(resultChan)

		// 参数验证
		if txHex == "" {
			log.ErrorWithContext(ctx, "解码原始交易失败：交易16进制字符串不能为空")
			resultChan <- AsyncResult{
				Result: nil,
				Error:  fmt.Errorf("交易16进制字符串不能为空"),
			}
			return
		}

		// 记录开始调用日志
		log.InfoWithContext(ctx, "开始解码原始交易", "txHexLength", len(txHex))

		// 调用区块链节点RPC（使用异步方式）
		asyncChan := CallRPCAsync(ctx, RpcMethodDecodeRawTransaction, []interface{}{txHex}, false)
		asyncResult := <-asyncChan

		if asyncResult.Error != nil {
			log.ErrorWithContext(ctx, "解码原始交易失败", "错误", asyncResult.Error)
			resultChan <- AsyncResult{
				Result: nil,
				Error:  fmt.Errorf("解码原始交易失败: %w", asyncResult.Error),
			}
			return
		}

		result := asyncResult.Result

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

		log.InfoWithContext(ctx, "成功解码原始交易")
		resultChan <- AsyncResult{
			Result: result,
			Error:  nil,
		}
	}()

	return resultChan
}

// FetchMemPoolTxs 获取内存池中的交易列表（异步）
func FetchMemPoolTxs(ctx context.Context) <-chan AsyncResult {
	resultChan := make(chan AsyncResult, 1)

	go func() {
		defer close(resultChan)

		log.InfoWithContext(ctx, "获取内存池交易列表")

		// 使用异步方式调用RPC
		asyncChan := CallRPCAsync(ctx, RpcMethodGetRawMempool, []interface{}{}, false)
		asyncResult := <-asyncChan

		if asyncResult.Error != nil {
			log.ErrorWithContext(ctx, "获取内存池交易列表失败", "error", asyncResult.Error)
			resultChan <- AsyncResult{
				Result: nil,
				Error:  asyncResult.Error,
			}
			return
		}

		response := asyncResult.Result

		// 验证响应数据类型
		txids, ok := response.([]interface{})
		if !ok {
			err := fmt.Errorf("响应格式错误")
			log.ErrorWithContext(ctx, "获取内存池交易列表响应格式错误")
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

		// 将txids转换为字符串数组
		txidList := make([]string, 0, len(txids))
		for _, txid := range txids {
			if txidStr, ok := txid.(string); ok {
				txidList = append(txidList, txidStr)
			}
		}

		// 构建响应结果
		result := map[string]interface{}{
			"tx_count": len(txidList),
			"txids":    txidList,
		}

		log.InfoWithContext(ctx, "获取内存池交易列表成功", "count", len(txidList))
		resultChan <- AsyncResult{
			Result: result,
			Error:  nil,
		}
	}()

	return resultChan
}

// GetTxVins 获取交易的输入数据
func GetTxVins(ctx context.Context, txids []string) <-chan AsyncResult {
	resultChan := make(chan AsyncResult, 1)

	go func() {
		defer close(resultChan)

		// 参数验证
		if len(txids) == 0 {
			log.ErrorWithContext(ctx, "获取交易输入数据失败：交易ID列表不能为空")
			resultChan <- AsyncResult{
				Result: nil,
				Error:  fmt.Errorf("交易ID列表不能为空"),
			}
			return
		}

		// 记录开始调用日志
		log.InfoWithContext(ctx, "开始获取交易输入数据", "txids", txids)

		// 存储处理结果
		result := make([]interface{}, 0, len(txids))

		for _, txid := range txids {
			// 1. 获取交易详情（使用异步方式）
			txAsyncChan := CallRPCAsync(ctx, RpcMethodGetRawTransaction, []interface{}{txid, 1}, false)
			txAsyncResult := <-txAsyncChan

			if txAsyncResult.Error != nil {
				log.ErrorWithContext(ctx, "获取交易详情失败", "txid", txid, "错误", txAsyncResult.Error)
				continue
			}

			txResp := txAsyncResult.Result

			txMap, ok := txResp.(map[string]interface{})
			if !ok {
				log.ErrorWithContext(ctx, "交易详情格式错误", "txid", txid)
				continue
			}

			// 获取交易的hash值
			hash, ok := txMap["hash"].(string)
			if !ok {
				log.ErrorWithContext(ctx, "获取交易hash失败", "txid", txid)
				continue
			}

			// 获取交易的输入列表
			vins, ok := txMap["vin"].([]interface{})
			if !ok {
				log.ErrorWithContext(ctx, "获取交易输入列表失败", "txid", txid)
				continue
			}

			vinDataList := []interface{}{}

			// 2. 处理每个输入
			for _, vinInterface := range vins {
				vin, ok := vinInterface.(map[string]interface{})
				if !ok {
					log.ErrorWithContext(ctx, "交易输入格式错误", "txid", txid)
					continue
				}

				// 检查是否为挖矿交易
				if coinbase, exists := vin["coinbase"]; exists {
					vinDataList = append(vinDataList, map[string]interface{}{
						"coinbase": coinbase,
					})
					continue
				}

				// 获取输入交易的ID
				vinTxid, ok := vin["txid"].(string)
				if !ok {
					log.ErrorWithContext(ctx, "获取输入交易ID失败", "txid", txid)
					continue
				}

				// 获取输入交易的原始数据（使用异步方式）
				vinRawAsyncChan := CallRPCAsync(ctx, RpcMethodGetRawTransaction, []interface{}{vinTxid}, false)
				vinRawAsyncResult := <-vinRawAsyncChan

				if vinRawAsyncResult.Error != nil {
					log.ErrorWithContext(ctx, "获取输入交易原始数据失败", "vinTxid", vinTxid, "错误", vinRawAsyncResult.Error)
					continue
				}

				vinRawResp := vinRawAsyncResult.Result

				vinRaw, ok := vinRawResp.(string)
				if !ok {
					log.ErrorWithContext(ctx, "输入交易原始数据格式错误", "vinTxid", vinTxid)
					continue
				}

				vinDataList = append(vinDataList, map[string]interface{}{
					"vin_txid": vinTxid,
					"vin_raw":  vinRaw,
				})
			}

			// 3. 添加到结果
			result = append(result, map[string]interface{}{
				"txid":     hash,
				"vin_data": vinDataList,
			})
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

		log.InfoWithContext(ctx, "成功获取交易输入数据", "count", len(result))
		resultChan <- AsyncResult{
			Result: result,
			Error:  nil,
		}
	}()

	return resultChan
}
