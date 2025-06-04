package ft

// FtHistoryRequest 获取FT交易历史的请求参数
type FtHistoryRequest struct {
	Address    string `uri:"address" binding:"required"`     // 用户地址
	ContractId string `uri:"contract_id" binding:"required"` // 代币合约ID
	Page       int    `uri:"page"`                           // 页码（从0开始）
	Size       int    `uri:"size"`                           // 每页记录数
}

// 验证请求参数是否合法
func (r *FtHistoryRequest) Validate() error {
	if r.Address == "" {
		return NewValidationError("地址不能为空")
	}
	if r.ContractId == "" {
		return NewValidationError("合约ID不能为空")
	}
	if r.Page < 0 {
		return NewValidationError("页码必须大于或等于0")
	}
	if r.Size < 1 || r.Size > 10000 {
		return NewValidationError("每页记录数必须在1到10000之间")
	}
	return nil
}

// FtHistoryRecord 单条FT交易历史记录
type FtHistoryRecord struct {
	TxId                   string   `json:"txid"`                     // 交易ID
	FtContractId           string   `json:"ft_contract_id"`           // 代币合约ID
	FtBalanceChange        int64    `json:"ft_balance_change"`        // 代币余额变化量
	FtDecimal              int      `json:"ft_decimal"`               // 代币精度
	TxFee                  float64  `json:"tx_fee"`                   // 交易费用
	SenderCombineScript    []string `json:"sender_combine_script"`    // 发送方地址列表
	RecipientCombineScript []string `json:"recipient_combine_script"` // 接收方地址列表
	TimeStamp              int64    `json:"time_stamp"`               // 交易时间戳
	UtcTime                string   `json:"utc_time"`                 // UTC时间字符串
}

// FtHistoryResponse FT交易历史响应
type FtHistoryResponse struct {
	Address      string            `json:"address"`       // 查询的地址
	ScriptHash   string            `json:"script_hash"`   // 生成的脚本哈希
	HistoryCount int               `json:"history_count"` // 历史记录总数
	Result       []FtHistoryRecord `json:"result"`        // 历史记录列表
}

// ValidationError 参数验证错误
type ValidationError struct {
	Message string
}

func (e ValidationError) Error() string {
	return e.Message
}

// NewValidationError 创建一个新的参数验证错误
func NewValidationError(message string) ValidationError {
	return ValidationError{Message: message}
}
