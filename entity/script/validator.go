package script

import (
	"fmt"
	"strings"
)

// ValidateScriptHash 验证脚本哈希是否合法
// 脚本哈希应该是一个64字符的十六进制字符串
func ValidateScriptHash(scriptHash string) error {
	// 检查长度
	if len(scriptHash) != 64 {
		return fmt.Errorf("脚本哈希长度必须为64个字符")
	}

	// 检查是否是十六进制
	for _, c := range scriptHash {
		if !strings.Contains("0123456789abcdefABCDEF", string(c)) {
			return fmt.Errorf("脚本哈希必须只包含十六进制字符")
		}
	}

	return nil
}
