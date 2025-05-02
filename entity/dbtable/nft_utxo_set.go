package dbtable

import (
	"gorm.io/gorm"
)

// NftUtxoSet 代表NFT UTXO集合表(nft_utxo_set)
type NftUtxoSet struct {
	// NFT合约ID，主键
	NftContractId string `gorm:"column:nft_contract_id;type:char(64);primaryKey;not null"`
	// 集合ID
	CollectionId string `gorm:"column:collection_id;type:char(64)"`
	// 集合索引
	CollectionIndex int `gorm:"column:collection_index;type:int"`
	// 集合名称
	CollectionName string `gorm:"column:collection_name;type:varchar(64)"`
	// NFT UTXO ID，唯一索引
	NftUtxoId string `gorm:"column:nft_utxo_id;type:char(64);uniqueIndex:nft_utxo_id;index:idx_utxo_id"`
	// NFT代码余额
	NftCodeBalance uint64 `gorm:"column:nft_code_balance;type:bigint unsigned"`
	// NFT P2PKH余额
	NftP2pkhBalance uint64 `gorm:"column:nft_p2pkh_balance;type:bigint unsigned"`
	// NFT名称
	NftName string `gorm:"column:nft_name;type:varchar(64)"`
	// NFT符号
	NftSymbol string `gorm:"column:nft_symbol;type:varchar(64)"`
	// NFT属性
	NftAttributes string `gorm:"column:nft_attributes;type:text"`
	// NFT描述
	NftDescription string `gorm:"column:nft_description;type:text"`
	// NFT转移次数
	NftTransferTimeCount int `gorm:"column:nft_transfer_time_count;type:int"`
	// NFT持有者地址
	NftHolderAddress string `gorm:"column:nft_holder_address;type:varchar(64)"`
	// NFT持有者脚本哈希
	NftHolderScriptHash string `gorm:"column:nft_holder_script_hash;type:char(64);index:idx_nft_holder"`
	// NFT创建时间戳
	NftCreateTimestamp int `gorm:"column:nft_create_timestamp;type:int"`
	// NFT最后转移时间戳
	NftLastTransferTimestamp int `gorm:"column:nft_last_transfer_timestamp;type:int"`
	// NFT图标
	NftIcon string `gorm:"column:nft_icon;type:mediumtext"`

	// gorm.Model的字段不包含在原始表中，但添加便于GORM管理
	gorm.Model `gorm:"-"` // 使用-标记表示该字段不存储到数据库
}

// TableName 指定表名为nft_utxo_set
func (NftUtxoSet) TableName() string {
	return "TBC20721.nft_utxo_set"
}
