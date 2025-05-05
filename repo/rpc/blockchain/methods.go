package blockchain

import (
	"context"
	"encoding/json"
	"fmt"

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

// GetBlockByHeightStructured 根据区块高度获取结构化区块信息
func GetBlockByHeightStructured(ctx context.Context, height int) (*BlockInfo, error) {
	// 参数验证
	if height < 0 {
		log.ErrorWithContext(ctx, "获取区块信息失败：区块高度不能为负数", "height:", height)
		return nil, fmt.Errorf("区块高度不能为负数")
	}

	// 记录开始调用日志
	log.InfoWithContext(ctx, "开始获取区块信息", "height:", height)

	// 调用区块链节点RPC
	result, err := CallRPC("getblockbyheight", []interface{}{height, true}, false)
	if err != nil {
		log.ErrorWithContext(ctx, "获取区块信息失败", "height:", height, "错误:", err)
		return nil, fmt.Errorf("获取区块信息失败: %w", err)
	}

	// 将返回结果转换为BlockInfo结构
	var blockInfo BlockInfo
	resultBytes, err := json.Marshal(result)
	if err != nil {
		log.ErrorWithContext(ctx, "序列化区块结果失败", "错误:", err)
		return nil, fmt.Errorf("解析RPC响应失败: %w", err)
	}

	if err := json.Unmarshal(resultBytes, &blockInfo); err != nil {
		log.ErrorWithContext(ctx, "解析区块数据失败", "height:", height, "错误:", err)
		return nil, fmt.Errorf("解析区块数据失败: %w", err)
	}

	log.InfoWithContext(ctx, "成功获取区块信息",
		"height:", height,
		"hash:", blockInfo.Hash,
		"time:", blockInfo.Time)
	return &blockInfo, nil
}

// FetchBlockByHeight 根据区块高度获取区块详情(原始接口数据)
func FetchBlockByHeight(ctx context.Context, height int64) (map[string]interface{}, error) {
	traceId := ctx.Value("trace_id")
	log.Info("[RPC请求] 通过高度获取区块", "trace_id", traceId, "height", height)

	response, err := CallRPC(RpcMethodGetBlockByHeight, []interface{}{height}, false)
	if err != nil {
		log.Error("[RPC请求失败] 通过高度获取区块", "trace_id", traceId, "height", height, "error", err)
		return nil, err
	}

	responseMap, ok := response.(map[string]interface{})
	if !ok {
		log.Error("[RPC响应格式错误] 通过高度获取区块", "trace_id", traceId, "height", height)
		return nil, fmt.Errorf("响应格式错误")
	}

	log.Info("[RPC请求成功] 通过高度获取区块", "trace_id", traceId, "height", height)
	return responseMap, nil
}

// FetchBlockByHash 根据区块哈希获取区块详情(原始接口数据)
func FetchBlockByHash(ctx context.Context, hash string) (map[string]interface{}, error) {
	traceId := ctx.Value("trace_id")
	log.Info("[RPC请求] 通过哈希获取区块", "trace_id", traceId, "hash", hash)

	response, err := CallRPC(RpcMethodGetBlock, []interface{}{hash}, false)
	if err != nil {
		log.Error("[RPC请求失败] 通过哈希获取区块", "trace_id", traceId, "hash", hash, "error", err)
		return nil, err
	}

	responseMap, ok := response.(map[string]interface{})
	if !ok {
		log.Error("[RPC响应格式错误] 通过哈希获取区块", "trace_id", traceId, "hash", hash)
		return nil, fmt.Errorf("响应格式错误")
	}

	log.Info("[RPC请求成功] 通过哈希获取区块", "trace_id", traceId, "hash", hash)
	return responseMap, nil
}

// FetchBlockHeaderByHeight 根据区块高度获取区块头信息
func FetchBlockHeaderByHeight(ctx context.Context, height int64) (map[string]interface{}, error) {
	traceId := ctx.Value("trace_id")
	log.Info("[RPC请求] 通过高度获取区块头", "trace_id", traceId, "height", height)

	// 先获取区块哈希
	hashResponse, err := CallRPC(RpcMethodGetBlockHash, []interface{}{height}, false)
	if err != nil {
		log.Error("[RPC请求失败] 获取区块哈希", "trace_id", traceId, "height", height, "error", err)
		return nil, err
	}

	hash, ok := hashResponse.(string)
	if !ok {
		log.Error("[RPC响应格式错误] 获取区块哈希", "trace_id", traceId, "height", height)
		return nil, fmt.Errorf("响应格式错误")
	}

	// 通过哈希获取区块头
	response, err := CallRPC(RpcMethodGetBlockHeader, []interface{}{hash}, false)
	if err != nil {
		log.Error("[RPC请求失败] 通过哈希获取区块头", "trace_id", traceId, "hash", hash, "error", err)
		return nil, err
	}

	responseMap, ok := response.(map[string]interface{})
	if !ok {
		log.Error("[RPC响应格式错误] 通过哈希获取区块头", "trace_id", traceId, "height", height)
		return nil, fmt.Errorf("响应格式错误")
	}

	log.Info("[RPC请求成功] 通过高度获取区块头", "trace_id", traceId, "height", height)
	return responseMap, nil
}

// FetchBlockHeaderByHash 根据区块哈希获取区块头信息
func FetchBlockHeaderByHash(ctx context.Context, hash string) (map[string]interface{}, error) {
	traceId := ctx.Value("trace_id")
	log.Info("[RPC请求] 通过哈希获取区块头", "trace_id", traceId, "hash", hash)

	response, err := CallRPC(RpcMethodGetBlockHeader, []interface{}{hash}, false)
	if err != nil {
		log.Error("[RPC请求失败] 通过哈希获取区块头", "trace_id", traceId, "hash", hash, "error", err)
		return nil, err
	}

	responseMap, ok := response.(map[string]interface{})
	if !ok {
		log.Error("[RPC响应格式错误] 通过哈希获取区块头", "trace_id", traceId, "hash", hash)
		return nil, fmt.Errorf("响应格式错误")
	}

	log.Info("[RPC请求成功] 通过哈希获取区块头", "trace_id", traceId, "hash", hash)
	return responseMap, nil
}

// FetchNearby10Headers 获取最近的10个区块头信息
func FetchNearby10Headers(ctx context.Context) ([]map[string]interface{}, error) {
	traceId := ctx.Value("trace_id")
	log.Info("[RPC请求] 获取最近10个区块头", "trace_id", traceId)

	// 获取当前区块高度
	infoResponse, err := CallRPC(RpcMethodGetInfo, []interface{}{}, false)
	if err != nil {
		log.Error("[RPC请求失败] 获取区块链信息", "trace_id", traceId, "error", err)
		return nil, err
	}

	info, ok := infoResponse.(map[string]interface{})
	if !ok {
		log.Error("[RPC响应格式错误] 获取区块链信息", "trace_id", traceId)
		return nil, fmt.Errorf("响应格式错误")
	}

	height, ok := info["blocks"].(float64)
	if !ok {
		log.Error("[RPC响应格式错误] 获取区块高度", "trace_id", traceId)
		return nil, fmt.Errorf("响应格式错误")
	}

	// 获取最近10个区块头
	var response []map[string]interface{}
	for i := 0; i < 10; i++ {
		blockHeight := int64(height) - int64(i)
		if blockHeight < 0 {
			break
		}

		header, err := FetchBlockHeaderByHeight(ctx, blockHeight)
		if err != nil {
			log.Error("[RPC请求失败] 获取区块头", "trace_id", traceId, "height", blockHeight, "error", err)
			continue
		}

		response = append(response, header)
	}

	log.Info("[RPC请求成功] 获取最近10个区块头", "trace_id", traceId, "count", len(response))
	return response, nil
}

// GetRawTransaction 获取交易原始信息
// 直接返回RPC调用结果，不进行结构体转换
func GetRawTransaction(ctx context.Context, txid string, verbose bool) (interface{}, error) {
	verboseInt := 0
	if verbose {
		verboseInt = 1
	}

	log.InfoWithContextf(ctx, "开始获取交易原始信息: txid=%s, verbose=%v", txid, verbose)
	result, err := CallRPC("getrawtransaction", []interface{}{txid, verboseInt}, false)
	if err != nil {
		log.ErrorWithContextf(ctx, "获取交易原始信息失败: txid=%s, 错误=%v", txid, err)
		return nil, fmt.Errorf("获取交易失败: %w", err)
	}

	log.InfoWithContextf(ctx, "成功获取交易原始信息: txid=%s", txid)
	return result, nil
}

// GetBlockByHeight 根据区块高度获取区块信息(简化版)
func GetBlockByHeight(ctx context.Context, height int64) (map[string]interface{}, error) {
	log.InfoWithContextf(ctx, "开始获取区块信息: height=%d", height)
	result, err := CallRPC("getblockbyheight", []interface{}{height}, false)
	if err != nil {
		log.ErrorWithContextf(ctx, "获取区块信息失败: height=%d, 错误=%v", height, err)
		return nil, fmt.Errorf("获取区块失败: %w", err)
	}

	resultMap, ok := result.(map[string]interface{})
	if !ok {
		log.ErrorWithContextf(ctx, "区块信息格式不正确: height=%d", height)
		return nil, fmt.Errorf("区块信息格式不正确")
	}

	log.InfoWithContextf(ctx, "成功获取区块信息: height=%d", height)
	return resultMap, nil
}

// FetchChainInfo 获取区块链信息
func FetchChainInfo(ctx context.Context) (map[string]interface{}, error) {
	traceId := ctx.Value("trace_id")
	log.Info("[RPC请求] 获取区块链信息", "trace_id", traceId)

	response, err := CallRPC(RpcMethodGetBlockchainInfo, []interface{}{}, false)
	if err != nil {
		log.Error("[RPC请求失败] 获取区块链信息", "trace_id", traceId, "error", err)
		return nil, err
	}

	responseMap, ok := response.(map[string]interface{})
	if !ok {
		log.Error("[RPC响应格式错误] 获取区块链信息", "trace_id", traceId)
		return nil, fmt.Errorf("响应格式错误")
	}

	log.Info("[RPC请求成功] 获取区块链信息", "trace_id", traceId)
	return responseMap, nil
}
