package electrumx

// ElectrumXHistoryItem 表示单个交易历史记录项
type ElectrumXHistoryItem struct {
    TxHash string `json:"tx_hash"`
    Height int    `json:"height"`
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
