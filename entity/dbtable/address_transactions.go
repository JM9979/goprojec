package dbtable

import (
	"time"
)

// AddressTransaction 地址交易关系表实体
type AddressTransaction struct {
	Fid           int64     `db:"Fid" gorm:"column:Fid;primaryKey"`
	Address       string    `db:"address" gorm:"column:address;index"`
	TxHash        string    `db:"tx_hash" gorm:"column:tx_hash;index"`
	IsSender      bool      `db:"is_sender" gorm:"column:is_sender"`
	IsRecipient   bool      `db:"is_recipient" gorm:"column:is_recipient"`
	BalanceChange float64   `db:"balance_change" gorm:"column:balance_change"`
	CreatedAt     time.Time `db:"created_at" gorm:"column:created_at;autoCreateTime"`
	UpdatedAt     time.Time `db:"updated_at" gorm:"column:updated_at;autoUpdateTime"`
}

// TableName 返回表名
func (AddressTransaction) TableName() string {
	return "TBC20721.address_transactions"
}

// AddressTransactionParam 查询参数
type AddressTransactionParam struct {
	Address string
	TxHash  string
	Offset  int
	Limit   int
}
