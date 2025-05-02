package ft

import (
	"fmt"
)

// FtInfoContractIdRequest 根据合约ID获取FT信息的请求
type FtInfoContractIdRequest struct {
	// FT合约ID
	ContractId string `uri:"contract_id" binding:"required"`
}

// Validate 验证FtInfoContractIdRequest的参数
func (req *FtInfoContractIdRequest) Validate() error {
	// 检查合约ID是否为空
	if req.ContractId == "" {
		return fmt.Errorf("合约ID不能为空")
	}

	// 检查合约ID格式是否合法（这里只是一个简单的示例，实际应用中可能需要更复杂的验证逻辑）
	if len(req.ContractId) < 8 {
		return fmt.Errorf("合约ID格式不正确")
	}

	return nil
}

// TBC20FTInfoResponse FT信息响应
type TBC20FTInfoResponse struct {
	FtContractId           string  `json:"ftContractId"`           // FT合约ID
	FtCodeScript           string  `json:"ftCodeScript"`           // FT代码脚本
	FtTapeScript           string  `json:"ftTapeScript"`           // FT磁带脚本
	FtSupply               float64 `json:"ftSupply"`               // FT总供应量（已考虑小数位）
	FtDecimal              int     `json:"ftDecimal"`              // FT小数位数
	FtName                 string  `json:"ftName"`                 // FT名称
	FtSymbol               string  `json:"ftSymbol"`               // FT符号
	FtDescription          string  `json:"ftDescription"`          // FT描述
	FtOriginUtxo           string  `json:"ftOriginUtxo"`           // FT起源UTXO
	FtCreatorCombineScript string  `json:"ftCreatorCombineScript"` // FT创建者的组合脚本
	FtHoldersCount         int     `json:"ftHoldersCount"`         // FT持有者数量
	FtIconUrl              string  `json:"ftIconUrl"`              // FT图标URL
	FtCreateTimestamp      int     `json:"ftCreateTimestamp"`      // FT创建时间戳
	FtTokenPrice           float64 `json:"ftTokenPrice"`           // FT代币价格
}
