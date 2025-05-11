package ft

import (
	"fmt"
)

// FtUtxoCombineScriptRequest 获取指定合并脚本和合约的FT UTXO请求
type FtUtxoCombineScriptRequest struct {
	// 合并脚本
	CombineScript string `uri:"combine_script" binding:"required"`
	// FT合约ID
	ContractId string `uri:"contract_id" binding:"required"`
}

/*
   "utxoId": "3843b853e78acd022f25ec52223a754e080d81529bb0cb345d594b679f202759",
   "utxoVout": 2,
   "utxoBalance": 500,
   "ftContractId": "a2d772d61afeac6b719a74d87872b9bbe847aa21b41a9473db066eabcddd86f3",
   "ftDecimal": 6,
   "ftBalance": 13897236
*/
// TBC20FTUtxoItem 单个TBC20 FT UTXO信息
type TBC20FTUtxoItem struct {
	UtxoId       string `json:"utxoId"`
	UtxoVout     int    `json:"utxoVout"`
	UtxoBalance  uint64 `json:"utxoBalance"`
	FtContractId string `json:"ftContractId"`
	FtDecimal    int    `json:"ftDecimal"`
	FtBalance    uint64 `json:"ftBalance"`
}

// TBC20FTUtxoResponse 获取指定合并脚本和合约的FT UTXO响应
type TBC20FTUtxoResponse struct {
	// FT UTXO列表
	Data []*TBC20FTUtxoItem `json:"data"`
}

// Validate 验证FtUtxoCombineScriptRequest的参数
func (req *FtUtxoCombineScriptRequest) Validate() error {
	// 检查合并脚本是否为空
	if req.CombineScript == "" {
		return fmt.Errorf("合并脚本不能为空")
	}

	// 检查合约ID是否为空
	if req.ContractId == "" {
		return fmt.Errorf("合约ID不能为空")
	}

	// 检查合并脚本格式是否合法
	if len(req.CombineScript) < 10 {
		return fmt.Errorf("合并脚本格式不正确")
	}

	// 检查合约ID格式是否合法
	if len(req.ContractId) < 8 {
		return fmt.Errorf("合约ID格式不正确")
	}

	return nil
}
