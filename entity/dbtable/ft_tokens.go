package dbtable

import (
	"gorm.io/gorm"
)

// FtTokens 代表功能代币表(ft_tokens)
type FtTokens struct {
	// 代币合约ID，主键
	FtContractId string `gorm:"column:ft_contract_id;type:char(64);primaryKey;not null"`
	// 代币代码脚本
	FtCodeScript string `gorm:"column:ft_code_script;type:text"`
	// 代币磁带脚本
	FtTapeScript string `gorm:"column:ft_tape_script;type:text"`
	// 代币供应量
	FtSupply uint64 `gorm:"column:ft_supply;type:bigint unsigned"`
	// 代币小数位数
	FtDecimal uint8 `gorm:"column:ft_decimal;type:tinyint unsigned"`
	// 代币名称
	FtName string `gorm:"column:ft_name;type:varchar(64);index:idx_name,length:20"`
	// 代币符号
	FtSymbol string `gorm:"column:ft_symbol;type:varchar(64)"`
	// 代币描述
	FtDescription string `gorm:"column:ft_description;type:text"`
	// 代币源UTXO
	FtOriginUtxo string `gorm:"column:ft_origin_utxo;type:char(72);uniqueIndex:ft_origin_utxo"`
	// 代币创建者组合脚本
	FtCreatorCombineScript string `gorm:"column:ft_creator_combine_script;type:char(42)"`
	// 代币持有者数量
	FtHoldersCount int `gorm:"column:ft_holders_count;type:int"`
	// 代币图标URL
	FtIconUrl string `gorm:"column:ft_icon_url;type:varchar(255)"`
	// 代币创建时间戳
	FtCreateTimestamp int `gorm:"column:ft_create_timestamp;type:int"`
	// 代币价格
	FtTokenPrice float64 `gorm:"column:ft_token_price;type:decimal(27,18)"`

	// gorm.Model的字段不包含在原始表中，但添加便于GORM管理
	gorm.Model `gorm:"-"` // 使用-标记表示该字段不存储到数据库
}

// TableName 指定表名为ft_tokens
func (FtTokens) TableName() string {
	return "TBC20721.ft_tokens"
}
