package electrumx

import (
	"context"
	"fmt"

	"ginproject/entity/electrumx"
	"ginproject/middleware/log"
)

// LocalUtxo 表示未花费交易输出
type LocalUtxo struct {
	TxHash string `json:"tx_hash"`
	TxPos  int    `json:"tx_pos"`
	Height int    `json:"height"`
	Value  int64  `json:"value"`
}

// GetTxHash 获取交易哈希
func (u LocalUtxo) GetTxHash() string {
	return u.TxHash
}

// ConvertUtxoResponse 将electrumx.UtxoResponse转换为自定义LocalUtxo切片
func ConvertUtxoResponse(response electrumx.UtxoResponse) []LocalUtxo {
	result := make([]LocalUtxo, 0, len(response))
	for _, item := range response {
		result = append(result, LocalUtxo{
			TxHash: item.TxHash,
			TxPos:  item.TxPos,
			Height: item.Height,
			Value:  item.Value,
		})
	}
	return result
}

// GetNftUtxos 获取NFT的UTXO列表
func GetNftUtxos(ctx context.Context, scriptHash string) ([]LocalUtxo, error) {
	// 参数校验
	if scriptHash == "" {
		log.ErrorWithContext(ctx, "获取NFT UTXO失败: scriptHash不能为空")
		return nil, fmt.Errorf("scriptHash不能为空")
	}

	// 记录开始调用日志
	log.InfoWithContext(ctx, "开始获取NFT UTXO",
		"scriptHash:", scriptHash)

	// 调用ElectrumX RPC获取未花费交易列表
	utxoResponse, err := GetUnspent(ctx, scriptHash)
	if err != nil {
		log.ErrorWithContext(ctx, "获取NFT UTXO失败",
			"scriptHash:", scriptHash,
			"错误:", err)
		return nil, fmt.Errorf("获取NFT UTXO失败: %w", err)
	}

	// 转换为LocalUtxo切片
	utxos := ConvertUtxoResponse(utxoResponse)

	log.InfoWithContext(ctx, "成功获取NFT UTXO",
		"scriptHash:", scriptHash,
		"count:", len(utxos))

	return utxos, nil
}
