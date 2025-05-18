package dbtable

import (
	"time"
)

// Role 参与角色类型
type Role string

const (
	RoleSender    Role = "sender"
	RoleRecipient Role = "recipient"
)

// TransactionParticipant 交易参与方信息表实体
type TransactionParticipant struct {
	Fid       int64     `db:"Fid" gorm:"column:Fid;primaryKey"`
	TxHash    string    `db:"tx_hash" gorm:"column:tx_hash;index"`
	Address   string    `db:"address" gorm:"column:address;index"`
	Role      Role      `db:"role" gorm:"column:role;type:enum('sender','recipient')"`
	CreatedAt time.Time `db:"created_at" gorm:"column:created_at;autoCreateTime"`
	UpdatedAt time.Time `db:"updated_at" gorm:"column:updated_at;autoUpdateTime"`
}

// TableName 返回表名
func (TransactionParticipant) TableName() string {
	return "TBC20721.transaction_participants"
}

// TransactionParticipantParam 查询参数
type TransactionParticipantParam struct {
	TxHash  string
	Address string
	Role    Role
	Offset  int
	Limit   int
}
