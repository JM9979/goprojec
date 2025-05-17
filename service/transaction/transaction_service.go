package transaction

import (
	"encoding/json"
	"net/http"

	txEntity "ginproject/entity/transaction"
	txLogic "ginproject/logic/transaction"
	"ginproject/middleware/log"

	"github.com/gin-gonic/gin"
)

// TransactionService 交易服务
type TransactionService struct{}

// NewTransactionService 创建新的交易服务实例
func NewTransactionService() *TransactionService {
	return &TransactionService{}
}

// BroadcastTxRaw 广播单笔原始交易
// POST /tx/raw
func (s *TransactionService) BroadcastTxRaw(c *gin.Context) {
	// 获取上下文
	ctx := c.Request.Context()

	// 解析请求参数
	var req txEntity.TxBroadcastRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.ErrorWithContext(ctx, "解析交易广播请求失败", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效: " + err.Error()})
		return
	}

	// 调用业务逻辑层处理请求
	resp, statusCode, err := txLogic.BroadcastTxRaw(ctx, &req)
	if err != nil {
		c.JSON(statusCode, gin.H{"error": err.Error()})
		return
	}

	// 返回结果
	c.JSON(statusCode, resp)
}

// DecodeTxRaw 解码原始交易
// POST /tx/raw/decode
func (s *TransactionService) DecodeTxRaw(c *gin.Context) {
	// 获取上下文
	ctx := c.Request.Context()

	// 解析请求参数
	var req txEntity.TxDecodeRawRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.ErrorWithContext(ctx, "解析交易解码请求失败", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效: " + err.Error()})
		return
	}

	// 调用业务逻辑层处理请求
	resp, statusCode, err := txLogic.DecodeRawTx(ctx, &req)
	if err != nil {
		c.JSON(statusCode, gin.H{"error": err.Error()})
		return
	}

	// 直接返回结果，不进行处理
	c.JSON(statusCode, resp)
}

// GetTxRawHex 获取交易原始十六进制数据
// GET /tx/hex/{txid}
func (s *TransactionService) GetTxRawHex(c *gin.Context) {
	// 获取上下文
	ctx := c.Request.Context()

	// 获取路径参数
	txid := c.Param("txid")
	if txid == "" {
		log.ErrorWithContext(ctx, "获取交易原始数据失败：交易ID不能为空")
		c.JSON(http.StatusBadRequest, gin.H{"error": "交易ID不能为空"})
		return
	}

	// 调用业务逻辑层处理请求
	txHex, statusCode, err := txLogic.GetTxRawHex(ctx, txid)
	if err != nil {
		c.JSON(statusCode, gin.H{"error": err.Error()})
		return
	}
	// json序列化
	btxHex, err := json.Marshal(txHex)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	// 返回结果
	c.String(statusCode, string(btxHex))
}

// DecodeTxByHash 通过交易ID解码交易
// GET /tx/hex/{txid}/decode
func (s *TransactionService) DecodeTxByHash(c *gin.Context) {
	// 获取上下文
	ctx := c.Request.Context()

	// 获取路径参数
	txid := c.Param("txid")
	if txid == "" {
		log.ErrorWithContext(ctx, "解码交易失败：交易ID不能为空")
		c.JSON(http.StatusBadRequest, gin.H{"error": "交易ID不能为空"})
		return
	}

	// 调用业务逻辑层处理请求
	resp, statusCode, err := txLogic.DecodeTxByHash(ctx, txid)
	if err != nil {
		c.JSON(statusCode, gin.H{"error": err.Error()})
		return
	}

	// 直接返回结果，不进行处理
	c.JSON(statusCode, resp)
}

// GetTxVins 获取交易输入数据
// POST /tx/vins
func (s *TransactionService) GetTxVins(c *gin.Context) {
	// 获取上下文
	ctx := c.Request.Context()

	// 解析请求参数
	var txids []string
	if err := c.ShouldBindJSON(&txids); err != nil {
		log.ErrorWithContext(ctx, "解析交易ID列表失败", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效: " + err.Error()})
		return
	}

	// 调用业务逻辑层处理请求
	resp, statusCode, err := txLogic.GetTxVins(ctx, txids)
	if err != nil {
		c.JSON(statusCode, gin.H{"error": err.Error()})
		return
	}

	// 返回结果
	c.JSON(statusCode, resp)
}
