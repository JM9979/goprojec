package http

import (
	"context"
	"testing"
	"time"
)

func TestMexcClient(t *testing.T) {
	// 创建上下文
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 创建客户端
	client := NewMexcClient()

	// 测试获取24小时行情
	symbol := "TBCUSDT"
	ticker24h, err := client.GetTicker24h(ctx, symbol)
	if err != nil {
		t.Fatalf("获取24小时行情失败: %v", err)
	}
	t.Logf("24小时行情: %+v", ticker24h)

	// 测试获取当前价格
	tickerPrice, err := client.GetTickerPrice(ctx, symbol)
	if err != nil {
		t.Fatalf("获取当前价格失败: %v", err)
	}
	t.Logf("当前价格: %+v", tickerPrice)
}

func TestMexcClient_GetTickerPrice(t *testing.T) {
	client := NewMexcClient()
	resp, err := client.GetTickerPrice(context.Background(), "TBCUSDT")
	if err != nil {
		t.Fatalf("获取MEXC当前价格失败: %v", err)
	}

	if resp == nil {
		t.Fatal("响应不应为空")
	}

	if resp.Symbol != "TBCUSDT" {
		t.Errorf("期望 Symbol=%s, 得到 %s", "TBCUSDT", resp.Symbol)
	}

	if resp.Price == "" {
		t.Error("价格不应为空")
	}
}
