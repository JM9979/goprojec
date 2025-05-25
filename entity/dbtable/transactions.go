package dbtable

import (
	"time"
)

// Transaction 交易主表实体
type Transaction struct {
	Fid       int64     `db:"Fid" gorm:"column:Fid;primaryKey"`
	TxHash    string    `db:"tx_hash" gorm:"column:tx_hash;uniqueIndex"`
	Fee       float64   `db:"fee" gorm:"column:fee"`
	TimeStamp int64     `db:"time_stamp" gorm:"column:time_stamp;index"`
	UtcTime   string    `db:"transaction_utc_time" gorm:"column:transaction_utc_time"`
	TxType    string    `db:"tx_type" gorm:"column:tx_type;index"`
	CreatedAt time.Time `db:"created_at" gorm:"column:created_at;autoCreateTime"`
	UpdatedAt time.Time `db:"updated_at" gorm:"column:updated_at;autoUpdateTime"`
}

// TableName 返回表名
func (Transaction) TableName() string {
	return "TBC20721.transactions"
}

// TransactionParam 查询参数
type TransactionParam struct {
	TxHash    string
	TimeStamp int64
	TxType    string
	Offset    int
	Limit     int
}
