package exchange

import (
	"context"
	"testing"
)

func TestGetExchangeRate(t *testing.T) {
	// 调用待测试的函数
	exchangeRate, err := GetExchangeRate(context.Background())

	// 检查是否有错误
	if err != nil {
		t.Fatalf("获取汇率信息失败: %v", err)
	}

	// 验证结果
	if exchangeRate == nil {
		t.Fatal("汇率信息不应为空")
	}

	// 验证货币单位
	if exchangeRate.Currency != "USD" {
		t.Errorf("期望货币单位为 USD，实际为 %s", exchangeRate.Currency)
	}

	// 验证时间戳
	if exchangeRate.Time <= 0 {
		t.Error("时间戳应大于 0")
	}

	// 记录测试结果
	t.Logf("获取到的汇率信息: %+v", exchangeRate)
}
