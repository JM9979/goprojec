package ft

import (
	"fmt"
)

// FtTokenHistoryRequest 获取代币历史交易记录请求参数
type FtTokenHistoryRequest struct {
	FtContractId string `uri:"ft_contract_id" binding:"required"` // 代币合约ID
	Page         int    `uri:"page"`           // 页码
	Size         int    `uri:"size"`           // 每页数量
}

// ValidateFtTokenHistoryRequest 验证获取代币历史交易记录请求参数
func ValidateFtTokenHistoryRequest(req *FtTokenHistoryRequest) error {
	// 检查代币合约ID不为空
	if req.FtContractId == "" {
		return fmt.Errorf("代币合约ID不能为空")
	}

	// 检查代币合约ID长度是否合法
	if len(req.FtContractId) != 64 {
		return fmt.Errorf("代币合约ID长度必须为64个字符")
	}

	// 检查代币合约ID是否为有效的十六进制字符
	for _, c := range req.FtContractId {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return fmt.Errorf("代币合约ID必须为有效的十六进制字符")
		}
	}

	// 检查分页参数
	if req.Page < 0 {
		return fmt.Errorf("页码不能小于0")
	}

	if req.Size <= 0 || req.Size > 10000 {
		return fmt.Errorf("每页数量必须在1到10000之间")
	}

	return nil
}

// FtTokenHistoryItem 代币历史交易记录项
type FtTokenHistoryItem struct {
	Txid         string              `json:"txid"`           // 交易ID
	FtContractId string              `json:"ft_contract_id"` // 代币合约ID
	TxInfo       *FtTxDecodeResponse `json:"tx_info"`        // 交易解码信息
}

// FtTokenHistoryResponse 获取代币历史交易记录响应
type FtTokenHistoryResponse []FtTokenHistoryItem
