package utility

import (
	"encoding/hex"
	"fmt"

	"ginproject/entity/constant"

	"github.com/btcsuite/btcutil/base58"
)

// APIResponse API通用响应结构
type APIResponse struct {
	// 错误码
	Code int `json:"code"`
	// 消息
	Message string `json:"message"`
	// 数据
	Data interface{} `json:"data"`
}

// NewSuccessResponse 创建成功响应
func NewSuccessResponse(data interface{}) APIResponse {
	return APIResponse{
		Code:    constant.CodeSuccess,
		Message: "success",
		Data:    data,
	}
}

// NewErrorResponse 创建错误响应
func NewErrorResponse(code int, message string) APIResponse {
	return APIResponse{
		Code:    code,
		Message: message,
		Data:    nil,
	}
}

// ConvertAddressToPublicKeyHash 将加密货币地址转换为公钥哈希
// 使用Base58解码，返回公钥哈希的十六进制字符串表示
func ConvertAddressToPublicKeyHash(address string) (string, error) {
	// Base58解码地址
	decoded, version, err := base58.CheckDecode(address)
	if err != nil {
		return "", fmt.Errorf("无效地址 '%s': %v", address, err)
	}

	// 记录版本号，但不使用
	_ = version

	// 直接使用解码后的数据作为公钥哈希
	// 注意：在某些区块链系统中，可能需要根据version处理不同类型的地址

	// 转换为十六进制字符串
	pubKeyHashHex := hex.EncodeToString(decoded)

	return pubKeyHashHex, nil
}
