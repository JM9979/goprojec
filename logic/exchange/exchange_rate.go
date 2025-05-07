package exchange

import (
	"context"
	"strconv"
	"time"

	"ginproject/entity/exchange"
	"ginproject/middleware/log"
	"ginproject/repo/http"
)

const (
	defaultSymbol  = "TBCUSDT"
	requestTimeout = 5 * time.Second
)

// GetExchangeRate 使用上下文获取TBC交易所汇率信息
func GetExchangeRate(ctx context.Context) (*exchange.ExchangeRateResponse, error) {
	log.InfoWithContext(ctx, "开始获取TBC交易所汇率信息")

	// 创建带超时的子上下文
	timeoutCtx, cancel := context.WithTimeout(ctx, requestTimeout)
	defer cancel()

	// 创建MEXC客户端
	client := http.NewMexcClient()

	// 直接在这里实现获取汇率的逻辑，不调用MexcClient.GetExchangeRate
	symbol := defaultSymbol
	log.InfoWithContext(ctx, "开始获取交易所汇率信息", "symbol", symbol)

	// 获取当前价格
	tickerPrice, err := client.GetTickerPrice(timeoutCtx, symbol)
	if err != nil {
		log.ErrorWithContext(ctx, "获取当前价格失败", "错误", err, "symbol", symbol)
		// 返回默认响应而不是中断
		return &exchange.ExchangeRateResponse{
			Currency: "USD",
			Rate:     0.0,
			Time:     time.Now().Unix(),
		}, nil
	}

	// 将价格字符串转换为浮点数
	rate, err := strconv.ParseFloat(tickerPrice.Price, 64)
	if err != nil {
		log.ErrorWithContext(ctx, "价格转换失败", "错误", err, "价格", tickerPrice.Price)
		rate = 0.0
	}

	// 获取24小时行情数据
	ticker24h, err := client.GetTicker24h(timeoutCtx, symbol)
	var changePercent string
	if err != nil {
		log.ErrorWithContext(ctx, "获取24小时行情数据失败", "错误", err, "symbol", symbol)
		// 不设置变化百分比
	} else {
		changePercent = ticker24h.PricePercent
	}

	currentTime := time.Now().Unix()
	exchangeRate := &exchange.ExchangeRateResponse{
		Currency:      "USD",
		Rate:          rate,
		Time:          currentTime,
		ChangePercent: changePercent,
	}

	log.InfoWithContext(ctx, "获取交易所汇率信息成功",
		"currency", exchangeRate.Currency,
		"rate", exchangeRate.Rate,
		"change_percent", exchangeRate.ChangePercent)

	return exchangeRate, nil
}
