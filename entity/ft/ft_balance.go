package ft

import (
	"fmt"

	"ginproject/entity/utility"
)

// FtBalanceAddressRequest 获取指定地址和合约的FT余额请求
type FtBalanceAddressRequest struct {
	// 用户钱包地址
	Address string `uri:"address" binding:"required"`
	// FT合约ID
	ContractId string `uri:"contract_id" binding:"required"`
}

// GetCombineScript 获取地址对应的组合脚本
func (req *FtBalanceAddressRequest) GetCombineScript() (string, error) {
	pubKeyHash, err := utility.ConvertAddressToPublicKeyHash(req.Address)
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

// FtBalanceMultiContractRequest 获取地址持有的多个代币余额请求
type FtBalanceMultiContractRequest struct {
	// 用户钱包地址
	Address string `uri:"address" binding:"required"`
	// FT合约ID列表
	FtContractId []string `json:"ftContractId" binding:"required"`
}

// GetCombineScript 获取地址对应的组合脚本
func (req *FtBalanceMultiContractRequest) GetCombineScript() (string, error) {
	pubKeyHash, err := utility.ConvertAddressToPublicKeyHash(req.Address)
	if err != nil {
		return "", err
	}
	return pubKeyHash, nil
}

// Validate 验证FtBalanceMultiContractRequest的参数
func (req *FtBalanceMultiContractRequest) Validate() error {
	// 检查地址是否为空
	if req.Address == "" {
		return fmt.Errorf("地址不能为空")
	}

	// 检查合约ID列表是否为空
	if len(req.FtContractId) == 0 {
		return fmt.Errorf("合约ID列表不能为空")
	}

	// 检查地址格式是否合法
	if len(req.Address) < 6 {
		return fmt.Errorf("地址格式不正确")
	}

	// 检查每个合约ID是否合法
	for _, contractId := range req.FtContractId {
		if len(contractId) < 8 {
			return fmt.Errorf("合约ID格式不正确: %s", contractId)
		}
	}

	return nil
}

// TBC20FTBalanceResponse 批量查询FT余额响应
type TBC20FTBalanceResponse struct {
	// 组合脚本（由地址转换而来）
	CombineScript string `json:"combineScript"`
	// FT合约ID
	FtContractId string `json:"ftContractId"`
	// FT小数位数
	FtDecimal int `json:"ftDecimal"`
	// FT余额
	FtBalance uint64 `json:"ftBalance"`
}

// TBC20TokenListHeldByCombineScriptRequest 通过合并脚本获取代币列表请求
type TBC20TokenListHeldByCombineScriptRequest struct {
	// 合并脚本
	CombineScript string `uri:"combine_script" binding:"required"`
}

// Validate 验证TBC20TokenListHeldByCombineScriptRequest的参数
func (req *TBC20TokenListHeldByCombineScriptRequest) Validate() error {
	// 检查合并脚本是否为空
	if req.CombineScript == "" {
		return fmt.Errorf("合并脚本不能为空")
	}

	return nil
}

// TBC20TokenListHeldByAddressRequest 获取地址持有的代币列表请求
type TBC20TokenListHeldByAddressRequest struct {
	// 用户钱包地址
	Address string `uri:"address" binding:"required"`
}

// GetCombineScript 获取地址对应的组合脚本
func (req *TBC20TokenListHeldByAddressRequest) GetCombineScript() (string, error) {
	pubKeyHash, err := utility.ConvertAddressToPublicKeyHash(req.Address)
	if err != nil {
		return "", err
	}
	return pubKeyHash, nil
}

// Validate 验证TBC20TokenListHeldByAddressRequest的参数
func (req *TBC20TokenListHeldByAddressRequest) Validate() error {
	// 检查地址是否为空
	if req.Address == "" {
		return fmt.Errorf("地址不能为空")
	}

	// 检查地址格式是否合法
	if len(req.Address) < 6 {
		return fmt.Errorf("地址格式不正确")
	}

	return nil
}

// TBC20TokenListHeldByAddressResponse 获取地址持有的代币列表响应
type TBC20TokenListHeldByAddressResponse struct {
	// 查询的地址
	Address string `json:"address"`
	// 地址持有的代币数量
	TokenCount int `json:"token_count"`
	// 代币列表
	TokenList []TokenInfo `json:"token_list"`
}

// TBC20TokenListHeldByCombineScriptResponse 通过合并脚本获取代币列表响应
type TBC20TokenListHeldByCombineScriptResponse struct {
	// 查询的地址
	CombineScript string `json:"combine_script"`
	// 地址持有的代币数量
	TokenCount int `json:"token_count"`
	// 代币列表
	TokenList []TokenInfo `json:"token_list"`
}

// TokenInfo 代币信息
type TokenInfo struct {
	// FT合约ID
	FtContractId string `json:"ft_contract_id"`
	// FT小数位数
	FtDecimal int `json:"ft_decimal"`
	// FT余额
	FtBalance uint64 `json:"ft_balance"`
	// FT名称
	FtName string `json:"ft_name"`
	// FT符号
	FtSymbol string `json:"ft_symbol"`
}
