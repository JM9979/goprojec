package broadcast

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
