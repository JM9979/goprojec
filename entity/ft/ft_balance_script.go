package ft

import (
	"fmt"
)

// FtBalanceCombineScriptRequest 获取指定合并脚本和合约的FT余额请求
type FtBalanceCombineScriptRequest struct {
	// 合并脚本
	CombineScript string `uri:"combine_script" binding:"required"`
	// 合约哈希
	ContractHash string `uri:"contract_hash" binding:"required"`
}

// FtBalanceCombineScriptResponse 
type FtBalanceCombineScriptResponse struct {
	// 合并脚本
	CombineScript string `json:"combineScript"`
	// 合约哈希
	ContractHash string `json:"ftContractId"`
	// 小数位数
	FtDecimal int `json:"ftDecimal"`
	// 余额
	FtBalance uint64 `json:"ftBalance"`
}

// Validate 验证FtBalanceCombineScriptRequest的参数
func (req *FtBalanceCombineScriptRequest) Validate() error {
	// A. 检查合并脚本是否为空
	if req.CombineScript == "" {
		return fmt.Errorf("合并脚本不能为空")
	}

	// B. 检查合约哈希是否为空
	if req.ContractHash == "" {
		return fmt.Errorf("合约哈希不能为空")
	}

	// C. 检查合并脚本格式是否合法
	if len(req.CombineScript) < 10 {
		return fmt.Errorf("合并脚本格式不正确")
	}

	// D. 检查合约哈希格式是否合法
	if len(req.ContractHash) < 8 {
		return fmt.Errorf("合约哈希格式不正确")
	}

	return nil
}
