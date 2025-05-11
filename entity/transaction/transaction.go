package transaction

import (
	"fmt"
	"strings"
)

// TxBroadcastRequest 单笔交易广播请求
type TxBroadcastRequest struct {
	TxHex string `json:"txHex" binding:"required"`
}

// TxsBroadcastRequest 批量交易广播请求
type TxsBroadcastRequest []TxBroadcastRequest

// BroadcastResponse 单笔交易广播响应
type BroadcastResponse struct {
	Result string          `json:"result,omitempty"`
	Error  *BroadcastError `json:"error,omitempty"`
}

// BroadcastError 广播错误信息
type BroadcastError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// InvalidTx 无效的交易
type InvalidTx struct {
	TxID         string `json:"txid"`
	RejectCode   int    `json:"reject_code"`
	RejectReason string `json:"reject_reason"`
}

// TxsBroadcastResponse 批量交易广播响应
type TxsBroadcastResponse struct {
	Result *TxsBroadcastResult `json:"result,omitempty"`
	Error  *BroadcastError     `json:"error,omitempty"`
}

// TxsBroadcastResult 批量交易广播结果
type TxsBroadcastResult struct {
	Invalid []InvalidTx `json:"invalid,omitempty"`
}

// TxDecodeRawRequest 解码原始交易请求
type TxDecodeRawRequest struct {
	TxHex string `json:"txHex" binding:"required"`
}

// ScriptSig 脚本签名
type ScriptSig struct {
	Asm string `json:"asm"`
	Hex string `json:"hex"`
}

// ScriptPubKey 公钥脚本
type ScriptPubKey struct {
	Asm       string   `json:"asm"`
	Hex       string   `json:"hex"`
	ReqSigs   int      `json:"reqSigs,omitempty"`
	Type      string   `json:"type"`
	Addresses []string `json:"addresses,omitempty"`
}

// Vin 交易输入
type Vin struct {
	TxID       string     `json:"txid,omitempty"`
	Vout       int        `json:"vout,omitempty"`
	ScriptSig  *ScriptSig `json:"scriptSig,omitempty"`
	Coinbase   string     `json:"coinbase,omitempty"`
	Sequence   int64      `json:"sequence"`
	Witness    []string   `json:"witness,omitempty"`
	IsCoinbase bool       `json:"-"` // 内部标记，不输出到JSON
}

// Vout 交易输出
type Vout struct {
	Value        float64       `json:"value"`
	N            int           `json:"n"`
	ScriptPubKey *ScriptPubKey `json:"scriptPubKey"`
}

// TxDecodeResponse 解码交易响应
type TxDecodeResponse struct {
	TxID          string `json:"txid"`
	Hash          string `json:"hash"`
	Version       int    `json:"version"`
	Size          int    `json:"size"`
	Locktime      int    `json:"locktime"`
	Vin           []Vin  `json:"vin"`
	Vout          []Vout `json:"vout"`
	BlockHash     string `json:"blockhash,omitempty"`
	Confirmations int    `json:"confirmations,omitempty"`
	Time          int64  `json:"time,omitempty"`
	BlockTime     int64  `json:"blocktime,omitempty"`
	BlockHeight   int    `json:"blockheight,omitempty"`
	Hex           string `json:"hex"`
}

// VinRawData 输入交易原始数据
type VinRawData struct {
	VinTxID string `json:"vin_txid"`
	VinRaw  string `json:"vin_raw"`
}

// CoinbaseVinData 挖矿交易输入数据
type CoinbaseVinData struct {
	Coinbase string `json:"coinbase"`
}

// TxVinsRawResponse 获取交易输入数据响应
type TxVinsRawResponse struct {
	TxID    string      `json:"txid"`
	VinData interface{} `json:"vin_data"`
}

// ValidateTxHex 验证交易16进制字符串
func ValidateTxHex(txHex string) error {
	if txHex == "" {
		return fmt.Errorf("交易16进制字符串不能为空")
	}

	// 验证是否为有效的16进制字符串
	txHex = strings.TrimSpace(txHex)
	for _, c := range txHex {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return fmt.Errorf("交易16进制字符串包含无效字符: %c", c)
		}
	}

	// 验证长度(至少应该有一个输入和一个输出)
	if len(txHex) < 20 {
		return fmt.Errorf("交易16进制字符串太短")
	}

	return nil
}

// Validate 验证交易广播请求
func (r *TxBroadcastRequest) Validate() error {
	return ValidateTxHex(r.TxHex)
}

// Validate 验证批量交易广播请求
func (r TxsBroadcastRequest) Validate() error {
	if len(r) == 0 {
		return fmt.Errorf("交易列表不能为空")
	}

	if len(r) > 100 {
		return fmt.Errorf("批量交易数量不能超过100个")
	}

	for i, tx := range r {
		if err := tx.Validate(); err != nil {
			return fmt.Errorf("第%d个交易校验失败: %w", i+1, err)
		}
	}

	return nil
}

// Validate 验证交易解码请求
func (r *TxDecodeRawRequest) Validate() error {
	return ValidateTxHex(r.TxHex)
}
