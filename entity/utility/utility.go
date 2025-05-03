package utility

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"

	"ginproject/entity/constant"

	"github.com/btcsuite/btcutil/base58"
	"golang.org/x/crypto/ripemd160"
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
	// 移除版本字节，只保留公钥哈希部分
	// if len(decoded) > 0 {
	// 	decoded = decoded[1:]
	// }
	// 转换为十六进制字符串
	pubKeyHashHex := hex.EncodeToString(decoded)

	return pubKeyHashHex, nil
}

// ConvertStrToSha256 将十六进制字符串转换为SHA256哈希值并以小端序返回
// 返回哈希值的十六进制字符串表示
func ConvertStrToSha256(input string) (string, error) {
	if input == "" {
		return "", fmt.Errorf("输入字符串不能为空")
	}

	// 将输入解析为十六进制字节
	inputBytes, err := hex.DecodeString(input)
	if err != nil {
		return "", fmt.Errorf("无效的十六进制字符串 '%s': %v", input, err)
	}

	// 计算SHA256哈希
	hashValue := sha256.Sum256(inputBytes)

	// 转换为小端序
	reversed := make([]byte, len(hashValue))
	for i := 0; i < len(hashValue); i++ {
		reversed[i] = hashValue[len(hashValue)-1-i]
	}

	// 转换为十六进制字符串
	hashHex := hex.EncodeToString(reversed)

	return hashHex, nil
}

// ConvertHexToSha256Reversed 将十六进制字符串转换为SHA256哈希并以小端序返回
// 输入：十六进制字符串
// 返回：小端序的SHA256哈希十六进制字符串
func ConvertHexToSha256Reversed(input string) (string, error) {

	// 将十六进制字符串解码为字节
	inputBytes, err := hex.DecodeString(input)
	if err != nil {
		return "", fmt.Errorf("无效的十六进制字符串 '%s': %v", input, err)
	}

	// 计算SHA256哈希
	hash := sha256.Sum256(inputBytes)

	// 转换为小端序
	reversed := make([]byte, len(hash))
	for i := 0; i < len(hash); i++ {
		reversed[i] = hash[len(hash)-1-i]
	}

	// 转换为十六进制字符串
	reversedHex := hex.EncodeToString(reversed)

	return reversedHex, nil
}

// ConvertCombineScriptToAddress 将组合脚本转换为地址
// 如果脚本以"00"结尾，转换为普通地址
// 否则返回带前缀的池控制或多重签名地址
func ConvertCombineScriptToAddress(combineScript string) (string, error) {
	if combineScript == "" {
		return "", fmt.Errorf("组合脚本不能为空")
	}

	// 检查组合脚本的长度是否至少为2
	if len(combineScript) < 2 {
		return "", fmt.Errorf("无效的组合脚本 '%s': 长度不足", combineScript)
	}

	// 判断是否为普通地址（以"00"结尾）
	if combineScript[len(combineScript)-2:] == "00" {
		// 普通地址
		pubKeyHash := combineScript[:len(combineScript)-2]
		// 将十六进制字符串转换为字节
		pubKeyHashBytes, err := hex.DecodeString(pubKeyHash)
		if err != nil {
			return "", fmt.Errorf("无效的公钥哈希 '%s': %v", pubKeyHash, err)
		}

		// 添加版本字节（0x00表示普通地址）并进行Base58Check编码
		version := byte(0x00)
		address := base58.CheckEncode(pubKeyHashBytes, version)
		return address, nil
	} else {
		// 池控制或多重签名地址
		return "Pool_or_ms_hash_" + combineScript, nil
	}
}

// ConvertMsAddressToMsScript 将多签地址转换为多签脚本
// 输入: 多签地址字符串
// 返回: 多签脚本字符串或错误
func ConvertMsAddressToMsScript(msAddress string) (string, error) {
	if msAddress == "" {
		return "", fmt.Errorf("多签地址不能为空")
	}

	// Base58解码地址
	decoded, version, err := base58.CheckDecode(msAddress)
	if err != nil {
		return "", fmt.Errorf("无效的多签地址 '%s': %v", msAddress, err)
	}

	// 从版本字节中提取签名需要的数量和总数量
	sigNeededCount := (version >> 4) & 0x0f
	sigTotalCount := version & 0x0f

	// 检查签名数量的合法性
	if sigNeededCount <= 0 || sigTotalCount <= 0 || sigNeededCount > sigTotalCount {
		return "", fmt.Errorf("无效的签名配置: 需要 %d, 总共 %d", sigNeededCount, sigTotalCount)
	}

	// 将解码后的数据作为公钥哈希，转换为十六进制字符串
	pubKeysHash := hex.EncodeToString(decoded)

	// 构建多签脚本
	msScript := fmt.Sprintf("%d PUB_KEYS %s CHECK_MULTISIG %d", sigNeededCount, pubKeysHash, sigTotalCount)

	return msScript, nil
}

// VerifyMsAddress 验证多签地址是否有效
// 输入: 多签地址字符串
// 返回: 是否有效，若无效则返回错误信息
func VerifyMsAddress(msAddress string) (bool, error) {
	if msAddress == "" {
		return false, fmt.Errorf("多签地址不能为空")
	}

	// Base58解码地址
	_, version, err := base58.CheckDecode(msAddress)
	if err != nil {
		return false, fmt.Errorf("无效的多签地址 '%s': %v", msAddress, err)
	}

	// 从版本字节中提取签名需要的数量和总数量
	sigNeededCount := (version >> 4) & 0x0f
	sigTotalCount := version & 0x0f

	// 验证签名数量的合法性
	if sigNeededCount <= 0 || sigTotalCount <= 0 || sigNeededCount > sigTotalCount {
		return false, fmt.Errorf("无效的签名配置: 需要 %d, 总共 %d", sigNeededCount, sigTotalCount)
	}

	return true, nil
}

// VerifyAMsAddress 验证多签地址是否有效（与Python版本保持一致的命名）
// 输入: 多签地址字符串
// 返回: 布尔值表示地址是否有效
func VerifyAMsAddress(msAddress string) bool {
	valid, _ := VerifyMsAddress(msAddress)
	return valid
}

// DigitalAsmToHex 将数字ASM转换为十六进制字符串
// 输入: ASM字符串（十进制数字）
// 返回: 字节序反转后的十六进制字符串，如果输入不是数字或超出4字节整数范围则直接返回输入
func DigitalAsmToHex(asm string) string {
	// 判断ASM是否是数字
	isDigit := true
	for _, c := range asm {
		if c < '0' || c > '9' {
			isDigit = false
			break
		}
	}

	if isDigit {
		// 将十进制字符串转换为整数
		var decimalValue uint64
		fmt.Sscanf(asm, "%d", &decimalValue)

		// 检查十进制值是否小于等于4字节整数的最大值（4294967295）
		if decimalValue <= 0xFFFFFFFF {
			// 转换为十六进制字符串，移除"0x"前缀
			hexValue := fmt.Sprintf("%x", decimalValue)

			// 确保十六进制值的长度是偶数，如果不是，添加前导零
			if len(hexValue)%2 != 0 {
				hexValue = "0" + hexValue
			}

			// 按字节反转（每两个字符反转）
			bytePairs := make([]string, 0, len(hexValue)/2)
			for i := 0; i < len(hexValue); i += 2 {
				end := i + 2
				if end > len(hexValue) {
					end = len(hexValue)
				}
				bytePairs = append(bytePairs, hexValue[i:end])
			}

			// 反转字节顺序
			for i, j := 0, len(bytePairs)-1; i < j; i, j = i+1, j-1 {
				bytePairs[i], bytePairs[j] = bytePairs[j], bytePairs[i]
			}

			return strings.Join(bytePairs, "")
		}
	}

	// 如果ASM不是数字，或者其值超过4字节，则直接返回ASM
	return asm
}

// GetPoolBalance 从tape_asm获取池余额
// 输入: tape_asm字符串
// 返回: FT LP余额、FT A余额和TBC余额
func GetPoolBalance(tapeAsm string) (int64, int64, int64, error) {
	if tapeAsm == "" {
		return 0, 0, 0, fmt.Errorf("tape_asm不能为空")
	}

	// 将tape_asm拆分为列表
	ftBalanceTapeList := strings.Split(tapeAsm, " ")

	// 检查列表长度是否足够
	if len(ftBalanceTapeList) < 4 {
		return 0, 0, 0, fmt.Errorf("无效的tape_asm格式: %s", tapeAsm)
	}

	complexBalance := ftBalanceTapeList[3]

	// 检查复杂余额字符串长度是否至少为48
	if len(complexBalance) < 48 {
		return 0, 0, 0, fmt.Errorf("复杂余额字符串长度不足: %s", complexBalance)
	}

	// 提取并解析FT LP余额（前16个字符，按2个字符反转）
	ftLpBalanceBytes := make([]string, 8)
	for i := 0; i < 16; i += 2 {
		ftLpBalanceBytes[i/2] = complexBalance[i : i+2]
	}
	// 反转字节顺序
	for i, j := 0, len(ftLpBalanceBytes)-1; i < j; i, j = i+1, j-1 {
		ftLpBalanceBytes[i], ftLpBalanceBytes[j] = ftLpBalanceBytes[j], ftLpBalanceBytes[i]
	}
	ftLpBalanceHex := strings.Join(ftLpBalanceBytes, "")
	var ftLpBalance int64
	fmt.Sscanf(ftLpBalanceHex, "%x", &ftLpBalance)

	// 提取并解析FT A余额（中间16个字符，按2个字符反转）
	ftABalanceBytes := make([]string, 8)
	for i := 16; i < 32; i += 2 {
		ftABalanceBytes[(i-16)/2] = complexBalance[i : i+2]
	}
	// 反转字节顺序
	for i, j := 0, len(ftABalanceBytes)-1; i < j; i, j = i+1, j-1 {
		ftABalanceBytes[i], ftABalanceBytes[j] = ftABalanceBytes[j], ftABalanceBytes[i]
	}
	ftABalanceHex := strings.Join(ftABalanceBytes, "")
	var ftABalance int64
	fmt.Sscanf(ftABalanceHex, "%x", &ftABalance)

	// 提取并解析TBC余额（末尾16个字符，按2个字符反转）
	tbcBalanceBytes := make([]string, 8)
	for i := 32; i < 48; i += 2 {
		tbcBalanceBytes[(i-32)/2] = complexBalance[i : i+2]
	}
	// 反转字节顺序
	for i, j := 0, len(tbcBalanceBytes)-1; i < j; i, j = i+1, j-1 {
		tbcBalanceBytes[i], tbcBalanceBytes[j] = tbcBalanceBytes[j], tbcBalanceBytes[i]
	}
	tbcBalanceHex := strings.Join(tbcBalanceBytes, "")
	var tbcBalance int64
	fmt.Sscanf(tbcBalanceHex, "%x", &tbcBalance)

	return ftLpBalance, ftABalance, tbcBalance, nil
}

// ConvertP2msUnlockScriptToAddress 将P2MS解锁脚本转换为地址
// 输入: P2MS解锁脚本字符串
// 返回: 地址字符串或错误
func ConvertP2msUnlockScriptToAddress(unlockScript string) (string, error) {
	if unlockScript == "" {
		return "", fmt.Errorf("解锁脚本不能为空")
	}

	// 将解锁脚本拆分为列表
	unlockScriptList := strings.Split(unlockScript, " ")

	// 检查脚本格式
	if len(unlockScriptList) < 3 {
		return "", fmt.Errorf("无效的解锁脚本格式: '%s'", unlockScript)
	}

	// 检查第一个元素是否为"0"
	if unlockScriptList[0] != "0" {
		return "", fmt.Errorf("无效的解锁脚本: 必须以'0'开始")
	}

	// 计算公钥需要数量和总数量
	pubkeyNeededCount := len(unlockScriptList) - 2

	// 计算公钥总数量 (基于最后一个元素长度除以66)
	lastElement := unlockScriptList[len(unlockScriptList)-1]
	pubkeyTotalCount := len(lastElement) / 66

	// 验证公钥数量的合法性
	if pubkeyNeededCount <= 0 || pubkeyTotalCount <= 0 || pubkeyNeededCount > pubkeyTotalCount || pubkeyTotalCount > 15 {
		return "", fmt.Errorf("无效的公钥配置: 需要 %d, 总共 %d", pubkeyNeededCount, pubkeyTotalCount)
	}

	// 计算版本字节 (pubkeyNeededCount << 4) | (pubkeyTotalCount & 0x0f)
	versionByte := byte((pubkeyNeededCount << 4) | (pubkeyTotalCount & 0x0f))

	// 将最后一个元素（公钥序列）转换为字节
	pubkeysHex := unlockScriptList[len(unlockScriptList)-1]
	pubkeysBytes, err := hex.DecodeString(pubkeysHex)
	if err != nil {
		return "", fmt.Errorf("无效的公钥十六进制数据: %v", err)
	}

	// 计算SHA256哈希
	sha256Hash := sha256.Sum256(pubkeysBytes)

	// 计算RIPEMD160哈希
	ripemd160Hasher := ripemd160.New()
	_, err = ripemd160Hasher.Write(sha256Hash[:])
	if err != nil {
		return "", fmt.Errorf("计算RIPEMD160哈希时出错: %v", err)
	}
	msPubkeysHash := ripemd160Hasher.Sum(nil)

	// 使用Base58Check编码
	msAddress := base58.CheckEncode(msPubkeysHash, versionByte)

	return msAddress, nil
}
