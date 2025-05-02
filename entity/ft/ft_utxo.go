package ft

import (
	"fmt"

	"ginproject/entity/utility"
)

// FtUtxoAddressRequest 获取指定地址和合约的FT UTXO请求
type FtUtxoAddressRequest struct {
	// 用户钱包地址
	Address string `uri:"address" binding:"required"`
	// FT合约ID
	ContractId string `uri:"contract_id" binding:"required"`
}

// GetCombineScript 获取地址对应的组合脚本
func (req *FtUtxoAddressRequest) GetCombineScript() (string, error) {
	pubKeyHash, err := utility.ConvertAddressToPublicKeyHash(req.Address)
	if err != nil {
		return "", err
	}
	return pubKeyHash, nil
}

// FtUtxoItem 单个FT UTXO信息
type FtUtxoItem struct {
	// 交易ID
	UtxoId string `json:"utxoId"`
	// 输出索引
	UtxoVout int `json:"utxoVout"`
	// UTXO余额
	UtxoBalance uint64 `json:"utxoBalance"`
	// FT合约ID
	FtContractId string `json:"ftContractId"`
	// FT小数位数
	FtDecimal int `json:"ftDecimal"`
	// FT余额
	FtBalance uint64 `json:"ftBalance"`
}

// FtUtxoAddressResponse 获取指定地址和合约的FT UTXO响应
type FtUtxoAddressResponse struct {
	// FT UTXO列表
	FtUtxoList []*FtUtxoItem `json:"ftUtxoList"`
}

// Validate 验证FtUtxoAddressRequest的参数
func (req *FtUtxoAddressRequest) Validate() error {
	// 检查地址是否为空
	if req.Address == "" {
		return fmt.Errorf("地址不能为空")
	}

	// 检查合约ID是否为空
	if req.ContractId == "" {
		return fmt.Errorf("合约ID不能为空")
	}

	// 检查地址格式是否合法（这里只是一个简单的示例，实际应用中可能需要更复杂的验证逻辑）
	if len(req.Address) < 6 {
		return fmt.Errorf("地址格式不正确")
	}

	// 检查合约ID格式是否合法（这里只是一个简单的示例，实际应用中可能需要更复杂的验证逻辑）
	if len(req.ContractId) < 8 {
		return fmt.Errorf("合约ID格式不正确")
	}

	return nil
}
