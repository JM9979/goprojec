package dbtable

import (
	"time"

	"gorm.io/gorm"
)

// FtTxHistory 代表功能代币交易历史表(ft_tx_history)
type FtTxHistory struct {
	// 自增ID
	ID uint `gorm:"column:id;primaryKey;autoIncrement"`
	// 交易ID
	Txid string `gorm:"column:txid;type:char(64);uniqueIndex;not null"`
	// 代币合约ID，外键关联ft_tokens表
	FtContractId string `gorm:"column:ft_contract_id;type:char(64);index"`
	// 持有者地址
	HolderAddress string `gorm:"column:holder_address;type:varchar(50);index"`
	// 脚本哈希
	ScriptHash string `gorm:"column:script_hash;type:char(64);index"`
	// 代币余额变化量
	FtBalanceChange int64 `gorm:"column:ft_balance_change;type:bigint"`
	// 交易费用
	TxFee float64 `gorm:"column:tx_fee;type:decimal(18,8)"`
	// 发送方地址列表，JSON格式
	SenderAddresses string `gorm:"column:sender_addresses;type:text"`
	// 接收方地址列表，JSON格式
	RecipientAddresses string `gorm:"column:recipient_addresses;type:text"`
	// 交易时间戳
	TimeStamp int64 `gorm:"column:time_stamp;type:bigint;index"`
	// UTC时间
	UtcTime string `gorm:"column:utc_time;type:varchar(30)"`
	// 确认状态
	Confirmed bool `gorm:"column:confirmed;type:tinyint(1);default:0"`
	// 创建时间
	CreatedAt time.Time `gorm:"column:created_at;type:timestamp;default:CURRENT_TIMESTAMP"`
	// 更新时间
	UpdatedAt time.Time `gorm:"column:updated_at;type:timestamp;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP"`
	// 删除时间
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;index"`
}

// TableName 指定表名为ft_tx_history
func (FtTxHistory) TableName() string {
	return "TBC20721.ft_tx_history"
}
