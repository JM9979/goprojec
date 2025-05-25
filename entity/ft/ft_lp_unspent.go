package ft

import (
	"fmt"
)

// LPUnspentByScriptHashRequest 获取LP未花费交易输出请求
type LPUnspentByScriptHashRequest struct {
	ScriptHash string `uri:"script_hash" binding:"required"`
}

// Validate 验证请求参数的合法性
func (req *LPUnspentByScriptHashRequest) Validate() error {
	// 检查脚本哈希是否为空
	if req.ScriptHash == "" {
		return fmt.Errorf("脚本哈希不能为空")
	}

	// 检查脚本哈希格式是否合法
	if len(req.ScriptHash) != 64 {
		return fmt.Errorf("脚本哈希格式不正确，应为64位十六进制字符串")
	}

	return nil
}

/*
{
    "ftUtxoList": [
        {
            "utxoId": "5b10dad299fda5f5b55c1f7a1a9a43cb82faf0a9d162d2c78f5490f304941c4f",
            "utxoVout": 4,
            "utxoBalance": 500,
            "ftContractId": "f6a012ef1cdef99706823b848c679a2d62964ac013212b0d6aafe3886225b828",
            "ftDecimal": null,
            "ftBalance": 884922
        }
    ]
}
*/

// TBC20FTLPUnspentResponse 获取LP未花费交易输出响应
type TBC20FTLPUnspentResponse struct {
	// 数据
	FtUtxoList []*TBC20FTLPUnspentItem `json:"ftUtxoList"`
}

// TBC20FTLPUnspentItem LP未花费交易输出信息
type TBC20FTLPUnspentItem struct {
	// 交易ID
	UtxoId string `json:"utxoId"`
	// 输出索引
	UtxoVout int `json:"utxoVout"`
	// UTXO余额
	UtxoBalance uint64 `json:"utxoBalance"`
	// FT合约ID
	FtContractId string `json:"ftContractId"`
	// FT小数位
	FtDecimal *int `json:"ftDecimal"`
	// FT余额
	FtBalance int64 `json:"ftBalance"`
}
