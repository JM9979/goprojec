package dbtable

import (
	"gorm.io/gorm"
)

// FtBalance 代表功能代币余额表(ft_balance)
type FtBalance struct {
	// 持有者组合脚本，主键之一
	FtHolderCombineScript string `gorm:"column:ft_holder_combine_script;type:char(42);primaryKey;not null"`
	// 代币合约ID，主键之一，外键关联ft_tokens表
	FtContractId string `gorm:"column:ft_contract_id;type:char(64);primaryKey;not null"`
	// 代币余额
	FtBalance uint64 `gorm:"column:ft_balance;type:bigint unsigned"`

	// gorm.Model的字段不包含在原始表中，但添加便于GORM管理
	gorm.Model `gorm:"-"` // 使用-标记表示该字段不存储到数据库
}

// TableName 指定表名为ft_balance
func (FtBalance) TableName() string {
	return "TBC20721.ft_balance"
}
