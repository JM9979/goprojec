package dbtable

import (
	"gorm.io/gorm"
)

// NftCollections 代表NFT集合表(nft_collections)
type NftCollections struct {
	// 集合ID，主键
	CollectionId string `gorm:"column:collection_id;type:char(64);primaryKey;not null"`
	// 集合名称
	CollectionName string `gorm:"column:collection_name;type:varchar(64)"`
	// 创建者地址
	CollectionCreatorAddress string `gorm:"column:collection_creator_address;type:varchar(64)"`
	// 创建者脚本哈希
	CollectionCreatorScriptHash string `gorm:"column:collection_creator_script_hash;type:char(64);index:idx_collection_creator"`
	// 集合符号
	CollectionSymbol string `gorm:"column:collection_symbol;type:varchar(64)"`
	// 集合属性
	CollectionAttributes string `gorm:"column:collection_attributes;type:text"`
	// 集合描述
	CollectionDescription string `gorm:"column:collection_description;type:text"`
	// 集合供应量
	CollectionSupply int `gorm:"column:collection_supply;type:int"`
	// 创建时间戳
	CollectionCreateTimestamp int `gorm:"column:collection_create_timestamp;type:int"`
	// 集合图标
	CollectionIcon string `gorm:"column:collection_icon;type:mediumtext"`

	// gorm.Model的字段不包含在原始表中，但添加便于GORM管理
	gorm.Model `gorm:"-"` // 使用-标记表示该字段不存储到数据库
}

// TableName 指定表名为nft_collections
func (NftCollections) TableName() string {
	return "TBC20721.nft_collections"
}
