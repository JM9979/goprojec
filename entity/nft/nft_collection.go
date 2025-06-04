package nft

// CollectionItem 表示单个NFT集合项目
type CollectionItem struct {
	CollectionId              string `json:"collectionId"`              // 集合ID
	CollectionName            string `json:"collectionName"`            // 集合名称
	CollectionCreator         string `json:"collectionCreator"`         // 集合创建者地址
	CollectionSymbol          string `json:"collectionSymbol"`          // 集合符号
	CollectionAttributes      string `json:"collectionAttributes"`      // 集合属性
	CollectionDescription     string `json:"collectionDescription"`     // 集合描述
	CollectionSupply          int    `json:"collectionSupply"`          // 集合供应量
	CollectionCreateTimestamp int    `json:"collectionCreateTimestamp"` // 创建时间戳
	CollectionIcon            string `json:"collectionIcon"`            // 集合图标
}

// CollectionDetailResponse 表示集合详情响应
type CollectionDetailResponse struct {
	CollectionId                string `json:"collectionId"`                // 集合ID
	CollectionName              string `json:"collectionName"`              // 集合名称
	CollectionCreator           string `json:"collectionCreator"`           // 集合创建者地址
	CollectionCreatorScripthash string `json:"collectionCreatorScripthash"` // 创建者脚本哈希
	CollectionSymbol            string `json:"collectionSymbol"`            // 集合符号
	CollectionAttributes        string `json:"collectionAttributes"`        // 集合属性
	CollectionDescription       string `json:"collectionDescription"`       // 集合描述
	CollectionSupply            int    `json:"collectionSupply"`            // 集合供应量
	CollectionCreateTimestamp   int    `json:"collectionCreateTimestamp"`   // 创建时间戳
	CollectionIcon              string `json:"collectionIcon"`              // 集合图标
}

// CollectionListResponse 表示集合列表响应
type CollectionListResponse struct {
	CollectionCount int              `json:"collectionCount"` // 集合总数
	CollectionList  []CollectionItem `json:"collectionList"`  // 集合列表
}

// CollectionQueryParams 表示查询NFT集合的参数
type CollectionQueryParams struct {
	Address string // 地址
	Page    int    // 页码
	Size    int    // 每页大小
}

// GetCollectionsPageSizeRequest 表示分页获取所有集合的请求参数
type GetCollectionsPageSizeRequest struct {
	Page int // 页码
	Size int // 每页大小
}

// GetDetailCollectionInfoRequest 表示获取集合详情的请求参数
type GetDetailCollectionInfoRequest struct {
	CollectionId string // 集合ID
}

// 集合相关错误定义
var (
	ErrEmptyCollectionAddress = NewNftError(20001, "集合查询地址不能为空")
	ErrInvalidCollectionPage  = NewNftError(20002, "集合查询页码不能为负数")
	ErrInvalidCollectionSize  = NewNftError(20003, "集合查询每页大小必须在1-10000之间")
	ErrEmptyCollectionId      = NewNftError(20004, "集合ID不能为空")
)

// ValidateCollectionQueryByAddress 验证按地址查询集合的参数
func ValidateCollectionQueryByAddress(address string, page, size int) error {
	// 验证地址
	if address == "" {
		return ErrEmptyCollectionAddress
	}

	// 验证页码
	if page < 0 {
		return ErrInvalidCollectionPage
	}

	// 验证每页大小
	if size <= 0 || size > 10000 {
		return ErrInvalidCollectionSize
	}

	return nil
}

// ValidateCollectionsPageSize 验证获取所有集合的分页参数
func ValidateCollectionsPageSize(page, size int) error {
	// 验证页码
	if page < 0 {
		return ErrInvalidCollectionPage
	}

	// 验证每页大小
	if size <= 0 || size > 10000 {
		return ErrInvalidCollectionSize
	}

	return nil
}

// ValidateDetailCollectionInfo 验证获取集合详情的参数
func ValidateDetailCollectionInfo(collectionId string) error {
	// 验证集合ID
	if collectionId == "" {
		return ErrEmptyCollectionId
	}

	return nil
}
