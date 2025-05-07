package exchange_service

import (
	"net/http"

	"ginproject/logic/exchange"
	"ginproject/middleware/log"

	"github.com/gin-gonic/gin"
)

// ExchangeService 提供交易所相关的API服务
type ExchangeService struct{}

// NewExchangeService 创建一个新的交易所服务实例
func NewExchangeService() *ExchangeService {
	return &ExchangeService{}
}

// GetExchangeRate 处理获取TBC交易所汇率信息的请求
func (s *ExchangeService) GetExchangeRate(c *gin.Context) {
	ctx := c.Request.Context()

	log.InfoWithContext(ctx, "收到获取TBC交易所汇率信息请求", "path", c.FullPath())

	// 调用逻辑层获取汇率信息，传递上下文
	exchangeRate, err := exchange.GetExchangeRate(ctx)
	if err != nil {
		log.ErrorWithContext(ctx, "获取TBC交易所汇率信息失败",
			"错误", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "获取汇率信息失败",
		})
		return
	}

	// 返回响应
	log.InfoWithContext(ctx, "成功获取TBC交易所汇率信息",
		"rate", exchangeRate.Rate,
		"change_percent", exchangeRate.ChangePercent)
	c.JSON(http.StatusOK, exchangeRate)
}
