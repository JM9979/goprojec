package address

import (
	"fmt"
	"ginproject/entity/electrumx"
	"regexp"
)

// 地址类型常量
const (
	AddressTypeInvalid = iota
	AddressTypeP2PKH
	AddressTypeP2SH
)

// AddressUnspentRequest 获取地址未花费UTXO的请求
type AddressUnspentRequest struct {
	// 用户钱包地址
	Address string `uri:"address" binding:"required"`
}

// Validate 验证地址参数的合法性
func (req *AddressUnspentRequest) Validate() error {
	valid, _, err := ValidateWIFAddress(req.Address)
	if !valid {
		return err
	}
	return nil
}

// ValidateWIFAddress 验证WIF格式的地址是否合法
func ValidateWIFAddress(address string) (bool, int, error) {
	// 检查地址是否为空
	if address == "" {
		return false, AddressTypeInvalid, fmt.Errorf("地址不能为空")
	}

	// 检查地址长度
	if len(address) < 26 || len(address) > 35 {
		return false, AddressTypeInvalid, fmt.Errorf("地址长度无效")
	}

	// 使用正则表达式验证基本格式
	// P2PKH地址通常以1开头
	// P2SH地址通常以3开头
	p2pkhRegex := regexp.MustCompile(`^1[a-km-zA-HJ-NP-Z1-9]{25,34}$`)
	p2shRegex := regexp.MustCompile(`^3[a-km-zA-HJ-NP-Z1-9]{25,34}$`)

	if p2pkhRegex.MatchString(address) {
		return true, AddressTypeP2PKH, nil
	}

	if p2shRegex.MatchString(address) {
		return true, AddressTypeP2SH, nil
	}

	return false, AddressTypeInvalid, fmt.Errorf("无效的地址格式")
}

// AddressUnspentResponse 获取地址未花费UTXO的响应
type AddressUnspentResponse struct {
	// UTXO列表
	Utxos electrumx.UtxoResponse `json:"utxos"`
}
