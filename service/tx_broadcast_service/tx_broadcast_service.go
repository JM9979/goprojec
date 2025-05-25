package tx_broadcast_service

import (
	"net/http"

	"ginproject/entity/broadcast"
	logic "ginproject/logic/broadcast"
	"ginproject/middleware/log"

	"github.com/gin-gonic/gin"
)

// TxBroadcastService 交易广播服务
type TxBroadcastService struct{}

// NewTxBroadcastService 创建新的交易广播服务实例
func NewTxBroadcastService() *TxBroadcastService {
	return &TxBroadcastService{}
}

// BroadcastTxRaw 广播单笔原始交易
func (s *TxBroadcastService) BroadcastTxRaw(c *gin.Context) {
	// 获取上下文
	ctx := c.Request.Context()

	// 解析请求参数
	var req broadcast.TxBroadcastRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.ErrorWithContext(ctx, "解析交易广播请求失败", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效: " + err.Error()})
		return
	}

	// 调用业务逻辑层处理请求
	resp, statusCode, err := logic.BroadcastTxRaw(ctx, &req)
	if err != nil {
		c.JSON(statusCode, gin.H{"error": err.Error()})
		return
	}

	// 返回结果
	c.JSON(statusCode, resp)
}

// BroadcastTxsRaw 批量广播原始交易
func (s *TxBroadcastService) BroadcastTxsRaw(c *gin.Context) {
	// 获取上下文
	ctx := c.Request.Context()

	// 解析请求参数
	var req broadcast.TxsBroadcastRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.ErrorWithContext(ctx, "解析批量交易广播请求失败", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效: " + err.Error()})
		return
	}

	// 调用业务逻辑层处理请求
	resp, statusCode, err := logic.BroadcastTxsRaw(ctx, req)
	if err != nil {
		c.JSON(statusCode, gin.H{"error": err.Error()})
		return
	}

	// 返回结果
	c.JSON(statusCode, resp)
}
