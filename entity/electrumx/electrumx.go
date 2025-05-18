package electrumx

// ElectrumXHistoryItem 表示单个交易历史记录项
type ElectrumXHistoryItem struct {
	TxHash string `json:"tx_hash"`
	Height int64    `json:"height"`
}

// ElectrumXHistoryResponse 表示从ElectrumX获取的历史记录响应
type ElectrumXHistoryResponse []ElectrumXHistoryItem

// Utxo 表示未花费交易输出
type Utxo struct {
	TxHash string `json:"tx_hash"` // 交易哈希
	TxPos  int    `json:"tx_pos"`  // 输出位置索引
	Height int    `json:"height"`  // 包含该交易的区块高度
	Value  int64  `json:"value"`   // UTXO金额（以聪为单位）
}

// UtxoResponse 表示从ElectrumX获取的UTXO响应
type UtxoResponse []Utxo

// AddressHistoryResponse 表示地址历史交易响应
type AddressHistoryResponse struct {
	Address      string        `json:"address"`       // 钱包地址
	Script       string        `json:"script"`        // 地址对应的脚本哈希
	HistoryCount int           `json:"history_count"` // 历史交易总数
	Result       []HistoryItem `json:"result"`        // 历史交易列表
}

// HistoryItem 表示单个历史交易记录
type HistoryItem struct {
	BalanceChange      string   `json:"balance_change"`       // 余额变动
	BanlanceChange     string   `json:"banlance_change"`      // 余额变动（字段重复，保持与接口规范一致）
	TxHash             string   `json:"tx_hash"`              // 交易哈希
	SenderAddresses    []string `json:"sender_addresses"`     // 发送方地址列表
	RecipientAddresses []string `json:"recipient_addresses"`  // 接收方地址列表
	Fee                string   `json:"fee"`                  // 交易费用
	TimeStamp          int64    `json:"time_stamp,omitempty"` // 交易时间戳（可选）
	UtcTime            string   `json:"utc_time"`             // UTC时间格式
	TxType             string   `json:"tx_type,omitempty"`    // 交易类型（可选）
}

// BalanceResponse 表示从ElectrumX获取的余额响应
type BalanceResponse struct {
	Confirmed   int64 `json:"confirmed"`   // 已确认的余额（以聪为单位）
	Unconfirmed int64 `json:"unconfirmed"` // 未确认的余额（以聪为单位）
}

// AddressBalanceResponse 表示格式化后的地址余额响应
type AddressBalanceResponse struct {
	Balance     int64 `json:"balance"`     // 总余额（已确认+未确认）
	Confirmed   int64 `json:"confirmed"`   // 已确认的余额
	Unconfirmed int64 `json:"unconfirmed"` // 未确认的余额
}

// FrozenBalanceResponse 表示冻结余额响应
type FrozenBalanceResponse struct {
	Frozen   int64 `json:"frozen_balance"`   // 冻结的余额（以聪为单位）
	// LockTime int64 `json:"locktime"` // 锁定时间戳
}
