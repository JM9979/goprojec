package blockchain

import (
	"encoding/hex"
	"errors"
	"strconv"
)

// DigitalAsmToHex 将数字ASM转换为十六进制
// 简化实现，保持原值
func DigitalAsmToHex(asmHex string) string {
	return asmHex
}

// ReverseHexToInt64 将反转的十六进制字符串转换为int64
func ReverseHexToInt64(hexStr string) (int64, error) {
	if len(hexStr) == 0 {
		return 0, errors.New("空的十六进制字符串")
	}

	// 每2个字符一组，反转顺序
	var reversed string
	for i := len(hexStr) - 2; i >= 0; i -= 2 {
		reversed += hexStr[i : i+2]
	}

	// 转换为int64
	return strconv.ParseInt(reversed, 16, 64)
}

// HexToString 将十六进制字符串转换为普通字符串
func HexToString(hexStr string) (string, error) {
	bytes, err := hex.DecodeString(hexStr)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// ReverseHexString 反转十六进制字符串，每2个字符作为一组
func ReverseHexString(hexStr string, startIndex, endIndex int) string {
	if startIndex < 0 || endIndex > len(hexStr) || startIndex >= endIndex {
		return ""
	}

	// 提取指定范围的子串
	subStr := hexStr[startIndex:endIndex]

	// 按每2个字符一组反转
	var result string
	for i := 0; i < len(subStr); i += 2 {
		if i+2 <= len(subStr) {
			result = subStr[i:i+2] + result
		} else {
			result = subStr[i:] + result
		}
	}

	return result
}

// HexToInt64 将十六进制字符串转换为int64
func HexToInt64(hexStr string) (int64, error) {
	if len(hexStr) == 0 {
		return 0, errors.New("空的十六进制字符串")
	}

	return strconv.ParseInt(hexStr, 16, 64)
}
