package tbcapi

import (
	"context"
	"encoding/hex"
	"errors"
	"strconv"
)

// BlockchainService 区块链服务接口
type BlockchainService struct {
}

// NewBlockchainService 创建区块链服务实例
func NewBlockchainService() *BlockchainService {
	return &BlockchainService{}
}

// DecodedTx 解码后的交易结构
type DecodedTx struct {
	Vout []struct {
		ScriptPubKey struct {
			Asm string
			Hex string
		}
	}
}

// DecodeTxHash 解码交易哈希，获取交易详情
func (s *BlockchainService) DecodeTxHash(ctx context.Context, txHash string) (*DecodedTx, error) {
	// 实际应调用区块链节点API获取交易信息
	// 这里使用模拟数据演示
	return s.mockDecodeTx(txHash), nil
}

// DigitalAsmToHex 将数字ASM转换为十六进制
func (s *BlockchainService) DigitalAsmToHex(asmHex string) string {
	// 简化处理，实际实现可能需要更复杂的逻辑
	return asmHex
}

// ReverseHexToInt64 将反转的十六进制字符串转换为int64
func (s *BlockchainService) ReverseHexToInt64(hexStr string) (int64, error) {
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
func (s *BlockchainService) HexToString(hexStr string) (string, error) {
	bytes, err := hex.DecodeString(hexStr)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// mockDecodeTx 模拟解码交易的结果
func (s *BlockchainService) mockDecodeTx(txHash string) *DecodedTx {
	// 模拟数据，实际应从区块链获取
	return &DecodedTx{
		Vout: []struct {
			ScriptPubKey struct {
				Asm string
				Hex string
			}
		}{
			{
				ScriptPubKey: struct {
					Asm string
					Hex string
				}{
					Asm: "OP_DUP OP_HASH160 a1b2c3d4e5f6 OP_EQUALVERIFY OP_CHECKSIG",
					Hex: "76a914a1b2c3d4e5f688ac",
				},
			},
			{
				ScriptPubKey: struct {
					Asm string
					Hex string
				}{
					Asm: "OP_RETURN 424950 a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6 203040102050607080900102030401020506070809001020304010205060708090 0123456789abcdef 3000",
					Hex: "6a424950a1b2c3d4e5f6...",
				},
			},
		},
	}
}
