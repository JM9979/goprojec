package utility

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"regexp"
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

// GetPoolBalanceFromTapeASM 从tape_asm获取池余额
// 输入: tape_asm字符串
// 返回: FT LP余额、FT A余额和TBC余额
func GetPoolBalanceFromTapeASM(tapeAsm string) (int64, int64, int64, error) {
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
	ftLpBalanceHex := reverseByteOrder(complexBalance[0:16])
	var ftLpBalance int64
	fmt.Sscanf(ftLpBalanceHex, "%x", &ftLpBalance)

	// 提取并解析FT A余额（中间16个字符，按2个字符反转）
	ftABalanceHex := reverseByteOrder(complexBalance[16:32])
	var ftABalance int64
	fmt.Sscanf(ftABalanceHex, "%x", &ftABalance)

	// 提取并解析TBC余额（末尾16个字符，按2个字符反转）
	tbcBalanceHex := reverseByteOrder(complexBalance[32:48])
	var tbcBalance int64
	fmt.Sscanf(tbcBalanceHex, "%x", &tbcBalance)

	return ftLpBalance, ftABalance, tbcBalance, nil
}

// reverseByteOrder 将十六进制字符串按2个字符为一组反转顺序
func reverseByteOrder(hexStr string) string {
	bytePairs := make([]string, 0, len(hexStr)/2)
	for i := 0; i < len(hexStr); i += 2 {
		end := i + 2
		if end > len(hexStr) {
			end = len(hexStr)
		}
		bytePairs = append(bytePairs, hexStr[i:end])
	}

	// 反转字节顺序
	for i, j := 0, len(bytePairs)-1; i < j; i, j = i+1, j-1 {
		bytePairs[i], bytePairs[j] = bytePairs[j], bytePairs[i]
	}

	return strings.Join(bytePairs, "")
}

// 地址类型常量
const (
	AddressTypeInvalid = iota
	AddressTypeP2PKH
	AddressTypeP2SH
)

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

// AddressToScriptHash 将比特币地址转换为ElectrumX使用的脚本哈希
// 首先将地址转换为公钥哈希，然后计算SHA256哈希，返回小端序的十六进制字符串
func AddressToScriptHash(address string) (string, error) {
	// 获取公钥哈希
	pubKeyHash, err := ConvertAddressToPublicKeyHash(address)
	if err != nil {
		return "", fmt.Errorf("转换地址到公钥哈希失败: %w", err)
	}

	// 添加脚本前缀和后缀
	// P2PKH脚本格式: OP_DUP OP_HASH160 <pubKeyHash> OP_EQUALVERIFY OP_CHECKSIG
	script := fmt.Sprintf("76a914%s88ac", pubKeyHash)

	// 将脚本转换为字节
	scriptBytes, err := hex.DecodeString(script)
	if err != nil {
		return "", fmt.Errorf("解码脚本失败: %w", err)
	}

	// 计算SHA256哈希
	hash := sha256.Sum256(scriptBytes)

	// 转换为小端序（反转字节顺序）
	reversed := make([]byte, len(hash))
	for i := 0; i < len(hash); i++ {
		reversed[i] = hash[len(hash)-1-i]
	}

	// 转换为十六进制字符串
	scriptHash := hex.EncodeToString(reversed)

	return scriptHash, nil
}

// ValidateAddress 验证比特币地址的格式合法性
func ValidateAddress(address string) (bool, error) {
	if address == "" {
		return false, fmt.Errorf("地址不能为空")
	}

	// 验证比特币地址的正则表达式
	// 支持P2PKH, P2SH和Bech32格式
	regexP2PKH := regexp.MustCompile(`^[13][a-km-zA-HJ-NP-Z1-9]{25,34}$`)
	regexP2SH := regexp.MustCompile(`^[2][a-km-zA-HJ-NP-Z1-9]{25,34}$`)
	regexBech32 := regexp.MustCompile(`^(bc1|tb1)[a-zA-HJ-NP-Z0-9]{25,90}$`)

	if regexP2PKH.MatchString(address) || regexP2SH.MatchString(address) || regexBech32.MatchString(address) {
		return true, nil
	}

	return false, fmt.Errorf("无效的比特币地址格式")
}

// ConvertP2msScriptToMsAddress 将P2MS脚本转换为多签名地址
// 输入: P2MS脚本字符串
// 返回: 多签名地址或错误
func ConvertP2msScriptToMsAddress(script string) (string, error) {
	if script == "" {
		return "", fmt.Errorf("脚本不能为空")
	}

	// 将脚本拆分为列表
	scriptParts := strings.Split(script, " ")

	// 验证脚本格式
	if len(scriptParts) < 3 {
		return "", fmt.Errorf("无效的脚本格式: '%s'", script)
	}

	// 查找 OP_CHECKMULTISIG 位置
	checkmultisigIndex := -1
	for i, part := range scriptParts {
		if part == "OP_CHECKMULTISIG" {
			checkmultisigIndex = i
			break
		}
	}

	if checkmultisigIndex == -1 {
		return "", fmt.Errorf("无效的多签名脚本: 缺少 OP_CHECKMULTISIG")
	}

	// 提取所需签名数量和公钥总数
	pubkeyNeededCount := 0
	if pubkeyNeededCountStr := scriptParts[0]; pubkeyNeededCountStr != "" {
		fmt.Sscanf(pubkeyNeededCountStr, "%d", &pubkeyNeededCount)
	}

	pubkeyTotalCount := 0
	if pubkeyTotalCountStr := scriptParts[checkmultisigIndex-1]; pubkeyTotalCountStr != "" {
		fmt.Sscanf(pubkeyTotalCountStr, "%d", &pubkeyTotalCount)
	}

	// 验证签名数量
	if pubkeyNeededCount <= 0 || pubkeyTotalCount <= 0 || pubkeyNeededCount > pubkeyTotalCount || pubkeyTotalCount > 15 {
		return "", fmt.Errorf("无效的签名配置: 需要 %d, 总共 %d", pubkeyNeededCount, pubkeyTotalCount)
	}

	// 提取公钥
	pubkeys := make([]string, 0, pubkeyTotalCount)
	for i := 1; i < checkmultisigIndex-1; i++ {
		// 验证是否为有效的公钥(通常为66字符的十六进制字符串)
		if len(scriptParts[i]) == 66 || len(scriptParts[i]) == 130 {
			pubkeys = append(pubkeys, scriptParts[i])
		}
	}

	// 验证提取的公钥数量
	if len(pubkeys) != pubkeyTotalCount {
		return "", fmt.Errorf("公钥数量不匹配: 提取到 %d, 期望 %d", len(pubkeys), pubkeyTotalCount)
	}

	// 连接所有公钥
	pubkeysStr := strings.Join(pubkeys, "")

	// 将公钥十六进制字符串转换为字节
	pubkeysBytes, err := hex.DecodeString(pubkeysStr)
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

	pubkeysHash := ripemd160Hasher.Sum(nil)

	// 计算版本字节 (pubkeyNeededCount << 4) | (pubkeyTotalCount & 0x0f)
	versionByte := byte((pubkeyNeededCount << 4) | (pubkeyTotalCount & 0x0f))

	// 使用Base58Check编码
	msAddress := base58.CheckEncode(pubkeysHash, versionByte)

	return msAddress, nil
}

// ConvertAddressToNftScriptHash 将加密货币地址转换为NFT脚本哈希
// 输入: 加密货币地址和是否为集合类型的标志
// 返回: 脚本哈希的十六进制字符串表示
func ConvertAddressToNftScriptHash(address string, isCollection bool) (string, error) {
	// 参数校验
	if address == "" {
		return "", fmt.Errorf("地址不能为空")
	}

	// 获取公钥哈希
	pubKeyHash, err := ConvertAddressToPublicKeyHash(address)
	if err != nil {
		return "", fmt.Errorf("转换地址到公钥哈希失败: %w", err)
	}

	// 公钥哈希转换为字节
	pubKeyHashBytes, err := hex.DecodeString(pubKeyHash)
	if err != nil {
		return "", fmt.Errorf("解码公钥哈希失败: %w", err)
	}

	// 构建脚本字节
	// P2PKH脚本部分: OP_DUP OP_HASH160 <pubKeyHash> OP_EQUALVERIFY OP_CHECKSIG
	script := []byte{0x76, 0xa9, 0x14}
	script = append(script, pubKeyHashBytes...)
	script = append(script, 0x88, 0xac)

	// 添加OP_RETURN数据
	// OP_RETURN (0x6a) 后跟数据长度 (0x0d = 13字节) 和数据
	script = append(script, 0x6a, 0x0d)

	// 根据isCollection参数添加不同的OP_RETURN数据
	if isCollection {
		// V0 Mint NHold
		script = append(script, []byte{0x56, 0x30, 0x20, 0x4d, 0x69, 0x6e, 0x74, 0x20, 0x4e, 0x48, 0x6f, 0x6c, 0x64}...)
	} else {
		// V0 Curr NHold
		script = append(script, []byte{0x56, 0x30, 0x20, 0x43, 0x75, 0x72, 0x72, 0x20, 0x4e, 0x48, 0x6f, 0x6c, 0x64}...)
	}

	// 计算SHA256哈希
	hash := sha256.Sum256(script)

	// 转换为小端序（反转字节顺序）
	reversed := make([]byte, len(hash))
	for i := 0; i < len(hash); i++ {
		reversed[i] = hash[len(hash)-1-i]
	}

	// 转换为十六进制字符串
	scriptHash := hex.EncodeToString(reversed)

	return scriptHash, nil
}

// HexToJson 将十六进制字符串转换为JSON对象
func HexToJson(hexStr string) (map[string]interface{}, error) {
	// 十六进制字符串转换为字节
	bytes, err := hex.DecodeString(hexStr)
	if err != nil {
		return nil, err
	}

	// 字节解析为JSON
	var result map[string]interface{}
	err = json.Unmarshal(bytes, &result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// ConvertCompressedPubkeyToLegacyAddress 将压缩公钥转换为传统地址
func ConvertCompressedPubkeyToLegacyAddress(pubkeyHex string) (string, error) {
	if len(pubkeyHex) != 66 {
		return "", fmt.Errorf("压缩公钥长度必须为66字符")
	}

	// 将十六进制字符串转换为字节
	pubkeyBytes, err := hex.DecodeString(pubkeyHex)
	if err != nil {
		return "", err
	}

	// 生成传统地址
	// 实际上这里需要根据项目的具体加密算法进行实现
	// 这里提供一个简化的实现

	// 简化实现，假设生成一个固定前缀的地址
	hash := sha256.Sum256(pubkeyBytes)
	ripemd160Hash := ripemd160.New()
	ripemd160Hash.Write(hash[:])
	pubKeyHash := ripemd160Hash.Sum(nil)

	// 添加版本前缀
	versionedPayload := append([]byte{0x00}, pubKeyHash...)

	// 计算校验和
	checksum := doubleSHA256(versionedPayload)[:4]

	// 组合最终结果
	fullPayload := append(versionedPayload, checksum...)

	// Base58编码
	address := base58.Encode(fullPayload)

	return address, nil
}

// doubleSHA256 辅助函数，计算double SHA256哈希
func doubleSHA256(data []byte) []byte {
	hash1 := sha256.Sum256(data)
	hash2 := sha256.Sum256(hash1[:])
	return hash2[:]
}
