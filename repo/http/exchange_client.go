package http

import (
	"context"
	"fmt"
	"net/url"

	"ginproject/entity/exchange"
	"ginproject/middleware/log"
)

const (
	// MEXC API的基本URL
	mexcBaseURL = "https://api.mexc.com"

	// API路径
	mexcTicker24hPath   = "/api/v3/ticker/24hr"
	mexcTickerPricePath = "/api/v3/ticker/price"
)

// MexcClient 提供MEXC交易所API的访问
type MexcClient struct {
	client *Client
}

// NewMexcClient 创建一个新的MEXC API客户端
func NewMexcClient() *MexcClient {
	return &MexcClient{
		client: NewClient(mexcBaseURL),
	}
}

// GetTicker24h 获取MEXC 24小时行情数据
func (m *MexcClient) GetTicker24h(ctx context.Context, symbol string) (*exchange.TickerInfo, error) {
	// 参数校验
	if symbol == "" {
		return nil, fmt.Errorf("symbol不能为空")
	}

	// 构建请求路径
	path := fmt.Sprintf("%s?%s", mexcTicker24hPath, url.Values{"symbol": {symbol}}.Encode())

	log.InfoWithContext(ctx, "获取MEXC 24小时行情数据", "symbol", symbol)
	resp, err := m.client.Get(ctx, path)
	if err != nil {
		log.ErrorWithContext(ctx, "获取MEXC 24小时行情数据失败", "错误", err, "symbol", symbol)
		return nil, fmt.Errorf("获取MEXC 24小时行情数据失败: %w", err)
	}

	var tickerInfo exchange.TickerInfo
	if err := ParseResponse(ctx, resp, &tickerInfo); err != nil {
		log.ErrorWithContext(ctx, "解析MEXC 24小时行情数据失败", "错误", err, "symbol", symbol)
		return nil, err
	}

	return &tickerInfo, nil
}

// GetTickerPrice 获取MEXC当前价格
func (m *MexcClient) GetTickerPrice(ctx context.Context, symbol string) (*exchange.TickerPrice, error) {
	// 参数校验
	if symbol == "" {
		return nil, fmt.Errorf("symbol不能为空")
	}

	// 构建请求路径
	path := fmt.Sprintf("%s?%s", mexcTickerPricePath, url.Values{"symbol": {symbol}}.Encode())

	log.InfoWithContext(ctx, "获取MEXC当前价格", "symbol", symbol)
	resp, err := m.client.Get(ctx, path)
	if err != nil {
		log.ErrorWithContext(ctx, "获取MEXC当前价格失败", "错误", err, "symbol", symbol)
		return nil, fmt.Errorf("获取MEXC当前价格失败: %w", err)
	}

	var tickerPrice exchange.TickerPrice
	if err := ParseResponse(ctx, resp, &tickerPrice); err != nil {
		log.ErrorWithContext(ctx, "解析MEXC当前价格失败", "错误", err, "symbol", symbol)
		return nil, err
	}

	return &tickerPrice, nil
}
