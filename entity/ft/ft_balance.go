package ft

import "fmt"

// FtBalanceAddressRequest 获取指定地址和合约的FT余额请求
type FtBalanceAddressRequest struct {
	// 用户钱包地址
	Address string `uri:"address" binding:"required"`
	// FT合约ID
	ContractId string `uri:"contract_id" binding:"required"`
}

// GetCombineScript 获取地址对应的组合脚本
func (req *FtBalanceAddressRequest) GetCombineScript() (string, error) {
	pubKeyHash, err := ConvertAddressToPublicKeyHash(req.Address)
	if err != nil {
		return "", err
	}
	return pubKeyHash, nil
}

// FtBalanceAddressResponse 获取指定地址和合约的FT余额响应
type FtBalanceAddressResponse struct {
	// 组合脚本（由地址转换而来）
	CombineScript string `json:"combineScript"`
	// FT合约ID
	FtContractId string `json:"ftContractId"`
	// FT小数位数
	FtDecimal int `json:"ftDecimal"`
	// FT余额
	FtBalance uint64 `json:"ftBalance"`
}

// Validate 验证FtBalanceAddressRequest的参数
func (req *FtBalanceAddressRequest) Validate() error {
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
