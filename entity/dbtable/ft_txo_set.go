package dbtable

import (
	"gorm.io/gorm"
)

// FtTxoSet 代表功能代币交易输出集合表(ft_txo_set)
type FtTxoSet struct {
	// UTXO交易ID，主键之一
	UtxoTxid string `gorm:"column:utxo_txid;type:char(64);primaryKey;not null"`
	// UTXO输出索引，主键之一
	UtxoVout int `gorm:"column:utxo_vout;primaryKey;not null"`
	// 持有者组合脚本哈希
	FtHolderCombineScript string `gorm:"column:ft_holder_combine_script;type:char(42)"`
	// 代币合约ID，外键关联ft_tokens表
	FtContractId string `gorm:"column:ft_contract_id;type:char(64);index:idx_script_hash_contract_id"`
	// UTXO余额
	UtxoBalance uint64 `gorm:"column:utxo_balance;type:bigint unsigned"`
	// 代币余额
	FtBalance uint64 `gorm:"column:ft_balance;type:bigint unsigned"`
	// 是否已花费
	IfSpend bool `gorm:"column:if_spend;type:tinyint(1);index:idx_if_spend"`

	// gorm.Model的字段不包含在原始表中，但添加便于GORM管理
	gorm.Model `gorm:"-"` // 使用-标记表示该字段不存储到数据库
}

// TableName 指定表名为ft_txo_set
func (FtTxoSet) TableName() string {
	return "TBC20721.ft_txo_set"
}
