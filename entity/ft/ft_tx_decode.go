package ft

import (
	"fmt"
)

// FtTxDecodeRequest 解析FT交易请求参数
type FtTxDecodeRequest struct {
	Txid string `uri:"txid" binding:"required"` // 交易ID
}

// ValidateFtTxDecodeRequest 验证解析FT交易请求参数
func ValidateFtTxDecodeRequest(req *FtTxDecodeRequest) error {
	// 检查交易ID不为空
	if req.Txid == "" {
		return fmt.Errorf("交易ID不能为空")
	}

	// 检查交易ID长度是否合法
	if len(req.Txid) != 64 {
		return fmt.Errorf("交易ID长度必须为64个字符")
	}

	// 检查交易ID是否为有效的十六进制字符
	for _, c := range req.Txid {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return fmt.Errorf("交易ID必须为有效的十六进制字符")
		}
	}

	return nil
}

// FtTxDecodeData 交易输入/输出数据项结构
type FtTxDecodeData struct {
	Txid       string `json:"txid"`        // 交易ID
	Vout       int    `json:"vout"`        // 输出索引
	Address    string `json:"address"`     // 地址
	ContractId string `json:"contract_id"` // 代币合约ID
	FtBalance  int64  `json:"ft_balance"`  // 代币数量
	FtDecimal  int    `json:"ft_decimal"`  // 代币小数位数
}

// FtTxDecodeResponse 解析FT交易响应
type FtTxDecodeResponse struct {
	Txid   string           `json:"txid"`   // 交易ID
	Input  []FtTxDecodeData `json:"input"`  // 交易输入数组
	Output []FtTxDecodeData `json:"output"` // 交易输出数组
}
