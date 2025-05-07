package nft

// NftItem 表示单个NFT项目
type NftItem struct {
	CollectionId          string `json:"collectionId"`          // 集合ID
	CollectionIndex       int    `json:"collectionIndex"`       // 集合索引
	CollectionName        string `json:"collectionName"`        // 集合名称
	CollectionIcon        string `json:"collectionIcon"`        // 集合图标
	CollectionDescription string `json:"collectionDescription"` // 集合描述
	NftContractId         string `json:"nftContractId"`         // NFT合约ID
	NftUtxoId             string `json:"nftUtxoId"`             // NFT UTXO ID
	NftCodeBalance        uint64 `json:"nftCodeBalance"`        // NFT代码余额
	NftP2pkhBalance       uint64 `json:"nftP2pkhBalance"`       // NFT P2PKH余额
	NftName               string `json:"nftName"`               // NFT名称
	NftSymbol             string `json:"nftSymbol"`             // NFT符号
	NftAttributes         string `json:"nftAttributes"`         // NFT属性
	NftDescription        string `json:"nftDescription"`        // NFT描述
	NftTransferTimeCount  int    `json:"nftTransferTimeCount"`  // NFT转移次数
	NftHolder             string `json:"nftHolder"`             // NFT持有者
	NftCreateTimestamp    int    `json:"nftCreateTimestamp"`    // NFT创建时间戳
	NftIcon               string `json:"nftIcon"`               // NFT图标
}

// NftListResponse 表示NFT列表响应
type NftListResponse struct {
	NftTotalCount int       `json:"nftTotalCount"` // NFT总数
	NftList       []NftItem `json:"nftList"`       // NFT列表
}

// NftInfoListResponse 表示NFT信息列表响应
type NftInfoListResponse struct {
	NftInfoList []NftItem `json:"nftInfoList"` // NFT信息列表
}

// 请求参数结构体
type NftsByAddressRequest struct {
	Address               string `json:"address" binding:"required"`                // 地址
	Page                  int    `json:"page" binding:"required"`                   // 页码
	Size                  int    `json:"size" binding:"required"`                   // 每页记录数
	IfExtraCollectionInfo bool   `json:"if_extra_collection_info_needed,omitempty"` // 是否需要额外的集合信息
}

// GetNftByAddressPageSizeRequest 表示按地址分页获取NFT的请求参数
type GetNftByAddressPageSizeRequest struct {
	Address               string // 地址
	Page                  int    // 页码
	Size                  int    // 每页记录数
	IfExtraCollectionInfo bool   // 是否需要额外的集合信息
}

// GetNftByScriptHashPageSizeRequest 表示按脚本哈希分页获取NFT的请求参数
type GetNftByScriptHashPageSizeRequest struct {
	ScriptHash string // 脚本哈希
	Page       int    // 页码
	Size       int    // 每页记录数
}

// GetNftByCollectionIdPageSizeRequest 表示按集合ID分页获取NFT的请求参数
type GetNftByCollectionIdPageSizeRequest struct {
	CollectionId string // 集合ID
	Page         int    // 页码
	Size         int    // 每页记录数
}

// GetNftsByContractIdsRequest 表示按合约ID列表获取NFT的请求参数
type GetNftsByContractIdsRequest struct {
	ContractList []string // 合约ID列表
	IfIconNeeded bool     // 是否需要图标
}

type NftsByScriptHashRequest struct {
	ScriptHash string `json:"script_hash" binding:"required"` // 脚本哈希
	Page       int    `json:"page" binding:"required"`        // 页码
	Size       int    `json:"size" binding:"required"`        // 每页记录数
}

type NftsByCollectionIdRequest struct {
	CollectionId string `json:"collection_id" binding:"required"` // 集合ID
	Page         int    `json:"page" binding:"required"`          // 页码
	Size         int    `json:"size" binding:"required"`          // 每页记录数
}

type CollectionsByAddressRequest struct {
	Address string `json:"address" binding:"required"` // 地址
	Page    int    `json:"page" binding:"required"`    // 页码
	Size    int    `json:"size" binding:"required"`    // 每页记录数
}

type CollectionsPageRequest struct {
	Page int `json:"page" binding:"required"` // 页码
	Size int `json:"size" binding:"required"` // 每页记录数
}

type NftsByContractIdsRequest struct {
	ContractList []string `json:"nft_contract_list" binding:"required"` // 合约ID列表
	IfIconNeeded bool     `json:"if_icon_needed,omitempty"`             // 是否需要图标
}

type DetailCollectionInfoRequest struct {
	CollectionId string `json:"collection_id" binding:"required"` // 集合ID
}

// 常量定义
const (
	MaxPageSize = 100
)

// 错误定义
var (
	ErrEmptyContractList = NewNftError(10008, "合约ID列表不能为空")
	ErrTooManyContracts  = NewNftError(10009, "合约ID列表不能超过100个")
)

// 验证方法

// ValidateGetNftsByContractIds 验证按合约ID列表获取NFT的请求参数
func ValidateGetNftsByContractIds(contractList []string, ifIconNeeded bool) error {
	if len(contractList) == 0 {
		return ErrEmptyContractList
	}
	if len(contractList) > MaxPageSize {
		return ErrTooManyContracts
	}
	return nil
}

// ValidateGetNftByAddressPageSize 验证按地址分页获取NFT的请求参数
func ValidateGetNftByAddressPageSize(address string, page, size int, ifExtraCollectionInfo bool) error {
	if address == "" {
		return ErrEmptyAddress
	}
	if page < 0 {
		return ErrInvalidPage
	}
	if size <= 0 || size > MaxPageSize {
		return ErrInvalidSize
	}
	return nil
}

// ValidateGetNftByScriptHashPageSize 验证按脚本哈希分页获取NFT的请求参数
func ValidateGetNftByScriptHashPageSize(scriptHash string, page, size int) error {
	if scriptHash == "" {
		return NewNftError(10004, "脚本哈希不能为空")
	}
	if page < 0 {
		return ErrInvalidPage
	}
	if size <= 0 || size > MaxPageSize {
		return ErrInvalidSize
	}
	return nil
}

// ValidateGetNftByCollectionIdPageSize 验证按集合ID分页获取NFT的请求参数
func ValidateGetNftByCollectionIdPageSize(collectionId string, page, size int) error {
	if collectionId == "" {
		return NewNftError(10005, "集合ID不能为空")
	}
	if page < 0 {
		return ErrInvalidPage
	}
	if size <= 0 || size > MaxPageSize {
		return ErrInvalidSize
	}
	return nil
}

// Validate 验证NftsByAddressRequest参数
func (r *NftsByAddressRequest) Validate() error {
	if r.Address == "" {
		return ErrEmptyAddress
	}
	if r.Page < 0 {
		return ErrInvalidPage
	}
	if r.Size <= 0 || r.Size > MaxPageSize {
		return ErrInvalidSize
	}
	return nil
}

// Validate 验证NftsByScriptHashRequest参数
func (r *NftsByScriptHashRequest) Validate() error {
	if r.ScriptHash == "" {
		return NewNftError(10004, "脚本哈希不能为空")
	}
	if r.Page < 0 {
		return ErrInvalidPage
	}
	if r.Size <= 0 || r.Size > MaxPageSize {
		return ErrInvalidSize
	}
	return nil
}

// Validate 验证NftsByCollectionIdRequest参数
func (r *NftsByCollectionIdRequest) Validate() error {
	if r.CollectionId == "" {
		return NewNftError(10005, "集合ID不能为空")
	}
	if r.Page < 0 {
		return ErrInvalidPage
	}
	if r.Size <= 0 || r.Size > MaxPageSize {
		return ErrInvalidSize
	}
	return nil
}

// Validate 验证CollectionsByAddressRequest参数
func (r *CollectionsByAddressRequest) Validate() error {
	if r.Address == "" {
		return ErrEmptyAddress
	}
	if r.Page < 0 {
		return ErrInvalidPage
	}
	if r.Size <= 0 || r.Size > MaxPageSize {
		return ErrInvalidSize
	}
	return nil
}

// Validate 验证CollectionsPageRequest参数
func (r *CollectionsPageRequest) Validate() error {
	if r.Page < 0 {
		return ErrInvalidPage
	}
	if r.Size <= 0 || r.Size > MaxPageSize {
		return ErrInvalidSize
	}
	return nil
}

// Validate 验证NftsByContractIdsRequest参数
func (r *NftsByContractIdsRequest) Validate() error {
	if len(r.ContractList) == 0 {
		return NewNftError(10006, "合约ID列表不能为空")
	}
	if len(r.ContractList) > MaxPageSize {
		return NewNftError(10007, "合约ID列表不能超过100个")
	}
	return nil
}

// Validate 验证DetailCollectionInfoRequest参数
func (r *DetailCollectionInfoRequest) Validate() error {
	if r.CollectionId == "" {
		return NewNftError(10005, "集合ID不能为空")
	}
	return nil
}
