package electrumx

// ElectrumXHistoryItem 表示单个交易历史记录项
type ElectrumXHistoryItem struct {
    TxHash string `json:"tx_hash"`
    Height int    `json:"height"`
}

// ElectrumXHistoryResponse 表示从ElectrumX获取的历史记录响应
type ElectrumXHistoryResponse []ElectrumXHistoryItem