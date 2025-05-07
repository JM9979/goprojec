package nft

// NftHistoryItem 表示NFT历史记录项
type NftHistoryItem struct {
	Txid               string   `json:"txid"`                // 交易ID
	CollectionId       string   `json:"collection_id"`       // NFT集合ID
	CollectionIndex    int      `json:"collection_index"`    // NFT集合索引
	CollectionName     string   `json:"collection_name"`     // NFT集合名称
	NftContractId      string   `json:"nft_contract_id"`     // NFT合约ID
	NftName            string   `json:"nft_name"`            // NFT名称
	NftSymbol          string   `json:"nft_symbol"`          // NFT符号
	NftDescription     string   `json:"nft_description"`     // NFT描述
	SenderAddresses    []string `json:"sender_addresses"`    // 发送方地址列表
	RecipientAddresses []string `json:"recipient_addresses"` // 接收方地址列表
	TimeStamp          *int64   `json:"time_stamp"`          // 交易时间戳
	UtcTime            string   `json:"utc_time"`            // UTC时间格式
	NftIcon            string   `json:"nft_icon"`            // NFT图标
}

// NftHistoryRequest 表示NFT历史记录请求参数
type NftHistoryRequest struct {
	Address string `json:"address" binding:"required"` // NFT持有者地址
	Page    int    `json:"page" binding:"required"`    // 页码
	Size    int    `json:"size" binding:"required"`    // 每页记录数
}

// NftHistoryResponse 表示NFT历史记录响应
type NftHistoryResponse struct {
	Address      string           `json:"address"`       // NFT持有者地址
	ScriptHash   string           `json:"script_hash"`   // 脚本哈希
	HistoryCount int              `json:"history_count"` // 历史记录总数
	Result       []NftHistoryItem `json:"result"`        // 历史记录列表
}

// AddressToNftScriptHashRequest 表示地址转换为NFT脚本哈希的请求参数
type AddressToNftScriptHashRequest struct {
	Address    string `json:"address" binding:"required"`       // 地址
	Collection bool   `json:"if_collection" binding:"required"` // 是否为集合
}

// 验证请求参数合法性
func (r *NftHistoryRequest) Validate() error {
	if r.Address == "" {
		return ErrEmptyAddress
	}
	if r.Page < 0 {
		return ErrInvalidPage
	}
	if r.Size <= 0 || r.Size > 100 {
		return ErrInvalidSize
	}
	return nil
}

// ValidateNftHistory 验证获取NFT历史记录的参数
func ValidateNftHistory(address string, page, size int) error {
	if address == "" {
		return ErrEmptyAddress
	}
	if page < 0 {
		return ErrInvalidPage
	}
	if size <= 0 || size > 100 {
		return ErrInvalidSize
	}
	return nil
}

// 错误定义
var (
	ErrEmptyAddress = NewNftError(10001, "地址不能为空")
	ErrInvalidPage  = NewNftError(10002, "页码不能为负数")
	ErrInvalidSize  = NewNftError(10003, "每页记录数必须在1-100之间")
)

// NftError 表示NFT操作相关的错误
type NftError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// Error 实现error接口
func (e *NftError) Error() string {
	return e.Message
}

// NewNftError 创建一个新的NFT错误
func NewNftError(code int, message string) *NftError {
	return &NftError{
		Code:    code,
		Message: message,
	}
}
