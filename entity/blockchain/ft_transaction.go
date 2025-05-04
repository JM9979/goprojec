package blockchain

// ScriptSig 表示脚本签名结构
type ScriptSig struct {
	Asm string `json:"asm"`
	Hex string `json:"hex"`
}

// ScriptPubKey 表示公钥脚本结构
type ScriptPubKey struct {
	Asm       string   `json:"asm"`
	Hex       string   `json:"hex"`
	ReqSigs   int      `json:"reqSigs,omitempty"`
	Type      string   `json:"type"`
	Addresses []string `json:"addresses,omitempty"`
}

// VinItem 表示交易输入项
type VinItem struct {
	Txid      string    `json:"txid"`
	Vout      int       `json:"vout"`
	ScriptSig ScriptSig `json:"scriptSig"`
	Sequence  int64     `json:"sequence"`
}

// VoutItem 表示交易输出项
type VoutItem struct {
	Value        float64      `json:"value"`
	N            int          `json:"n"`
	ScriptPubKey ScriptPubKey `json:"scriptPubKey"`
}

// TransactionResponse 表示交易信息的响应结构
type TransactionResponse struct {
	Txid          string     `json:"txid"`
	Hash          string     `json:"hash"`
	Version       int        `json:"version"`
	Size          int        `json:"size"`
	Vsize         int        `json:"vsize"`
	Weight        int        `json:"weight"`
	LockTime      int        `json:"locktime"`
	Vin           []VinItem  `json:"vin"`
	Vout          []VoutItem `json:"vout"`
	Hex           string     `json:"hex"`
	Blockhash     string     `json:"blockhash,omitempty"`
	Confirmations int        `json:"confirmations,omitempty"`
	Time          int64      `json:"time,omitempty"`
	Blocktime     int64      `json:"blocktime,omitempty"`
}
