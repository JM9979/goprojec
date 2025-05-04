package blockchain

import (
	"context"
	"encoding/json"
	"fmt"

	"ginproject/middleware/log"

	"go.uber.org/zap"
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

// GetBlockByHeight 根据区块高度获取区块信息
func GetBlockByHeight(ctx context.Context, height int) (*BlockInfo, error) {
	// 参数验证
	if height < 0 {
		log.ErrorWithContext(ctx, "获取区块信息失败：区块高度不能为负数", zap.Int("height", height))
		return nil, fmt.Errorf("区块高度不能为负数")
	}

	// 记录开始调用日志
	log.InfoWithContext(ctx, "开始获取区块信息", zap.Int("height", height))

	// 调用区块链节点RPC
	result, err := CallRPC("getblockbyheight", []interface{}{height, true}, false)
	if err != nil {
		log.ErrorWithContext(ctx, "获取区块信息失败", zap.Int("height", height), zap.Error(err))
		return nil, fmt.Errorf("获取区块信息失败: %w", err)
	}

	// 将返回结果转换为BlockInfo结构
	var blockInfo BlockInfo
	resultBytes, err := json.Marshal(result)
	if err != nil {
		log.ErrorWithContext(ctx, "序列化区块结果失败", zap.Error(err))
		return nil, fmt.Errorf("解析RPC响应失败: %w", err)
	}

	if err := json.Unmarshal(resultBytes, &blockInfo); err != nil {
		log.ErrorWithContext(ctx, "解析区块数据失败", zap.Int("height", height), zap.Error(err))
		return nil, fmt.Errorf("解析区块数据失败: %w", err)
	}

	log.InfoWithContext(ctx, "成功获取区块信息",
		zap.Int("height", height),
		zap.String("hash", blockInfo.Hash),
		zap.Int64("time", blockInfo.Time))
	return &blockInfo, nil
}
