package exchange

// TickerInfo 表示24小时行情数据
type TickerInfo struct {
	Symbol       string `json:"symbol"`
	PriceChange  string `json:"priceChange"`
	PricePercent string `json:"priceChangePercent"`
	LastPrice    string `json:"lastPrice"`
	OpenPrice    string `json:"openPrice"`
	HighPrice    string `json:"highPrice"`
	LowPrice     string `json:"lowPrice"`
	Volume       string `json:"volume"`
	QuoteVolume  string `json:"quoteVolume"`
	OpenTime     int64  `json:"openTime"`
	CloseTime    int64  `json:"closeTime"`
	Count        int64  `json:"count"`
}

// TickerPrice 表示当前价格数据
type TickerPrice struct {
	Symbol string `json:"symbol"`
	Price  string `json:"price"`
}

// ExchangeRateResponse 表示汇率响应数据
type ExchangeRateResponse struct {
	Currency      string  `json:"currency,omitempty"`       // 货币单位，默认为"USD"
	Rate          float64 `json:"rate,omitempty"`           // TBC对USD的汇率
	Time          int64   `json:"time,omitempty"`           // 时间戳（Unix格式）
	ChangePercent string  `json:"change_percent,omitempty"` // 24小时价格变化百分比
}

// ExchangeType 表示交易所类型
type ExchangeType string

const (
	// MexcExchange 表示MEXC交易所
	MexcExchange ExchangeType = "mexc"
	// 将来可以添加更多交易所类型
)
