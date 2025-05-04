package addressservice

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"ginproject/logic/address"
	"ginproject/middleware/log"
)

// AddressService 地址服务
type AddressService struct {
	addressLogic *address.AddressLogic
}

// NewAddressService 创建地址服务实例
func NewAddressService() *AddressService {
	return &AddressService{
		addressLogic: address.NewAddressLogic(),
	}
}

// GetAddressUnspentUtxos 获取地址未花费交易输出(UTXO)
// @Router /v1/tbc/main/address/{address}/unspent/ [get]
func (s *AddressService) GetAddressUnspentUtxos(c *gin.Context) {
	// 获取上下文和参数
	ctx := c.Request.Context()
	address := c.Param("address")

	// 记录请求日志
	log.InfoWithContext(ctx, "收到获取地址UTXO请求")

	// 参数验证
	if address == "" {
		c.JSON(http.StatusOK, gin.H{
			"code":    http.StatusBadRequest,
			"message": "地址参数不能为空",
		})
		return
	}

	// 调用业务逻辑层
	utxos, err := s.addressLogic.GetAddressUnspentUtxos(ctx, address)
	if err != nil {
		log.ErrorWithContext(ctx, "获取地址UTXO失败", zap.Error(err))
		c.JSON(http.StatusOK, gin.H{
			"code":    http.StatusInternalServerError,
			"message": "获取地址UTXO失败: " + err.Error(),
		})
		return
	}

	// 返回成功响应
	c.JSON(http.StatusOK, utxos)
}

// GetAddressHistory 获取地址历史交易信息
// @Router /v1/tbc/main/address/{address}/history [get]
func (s *AddressService) GetAddressHistory(c *gin.Context) {
	// 获取上下文和参数
	ctx := c.Request.Context()
	address := c.Param("address")

	// 记录请求日志
	log.InfoWithContext(ctx, "收到获取地址历史交易请求", zap.String("address", address))

	// 参数验证
	if address == "" {
		c.JSON(http.StatusOK, gin.H{
			"code":    http.StatusBadRequest,
			"message": "地址参数不能为空",
		})
		return
	}

	// 调用业务逻辑层 - 使用默认模式，返回最新30条交易
	history, err := s.addressLogic.GetAddressHistoryPage(ctx, address, false, 0)
	if err != nil {
		log.ErrorWithContextf(ctx, "获取地址历史交易失败: %v", err)
		c.JSON(http.StatusOK, gin.H{
			"code":    http.StatusInternalServerError,
			"message": "获取地址历史交易失败: " + err.Error(),
		})
		return
	}

	// 返回成功响应
	c.JSON(http.StatusOK, history)
}

// GetAddressHistoryPaged 获取地址历史交易信息（分页版）
// @Router /v1/tbc/main/address/{address}/history/page/{page} [get]
func (s *AddressService) GetAddressHistoryPaged(c *gin.Context) {
	// 获取上下文和参数
	ctx := c.Request.Context()
	address := c.Param("address")

	// 获取分页参数
	pageStr := c.DefaultQuery("page", "0")
	page := 0
	if pageStr != "" {
		page, _ = strconv.Atoi(pageStr)
	}

	// 记录请求日志
	log.InfoWithContext(ctx, "收到获取地址历史交易分页请求",
		zap.String("address", address),
		zap.Int("page", page))

	// 参数验证
	if address == "" {
		c.JSON(http.StatusOK, gin.H{
			"code":    http.StatusBadRequest,
			"message": "地址参数不能为空",
		})
		return
	}

	// 调用业务逻辑层，启用分页模式
	history, err := s.addressLogic.GetAddressHistoryPage(ctx, address, true, page)
	if err != nil {
		log.ErrorWithContextf(ctx, "获取地址历史交易失败: %v", err)
		c.JSON(http.StatusOK, gin.H{
			"code":    http.StatusInternalServerError,
			"message": "获取地址历史交易失败: " + err.Error(),
		})
		return
	}

	// 返回成功响应
	c.JSON(http.StatusOK, history)
}

// GetAddressBalance 获取地址余额
// @Router /v1/tbc/main/address/{address}/get/balance [get]
func (s *AddressService) GetAddressBalance(c *gin.Context) {
	// 获取上下文和参数
	ctx := c.Request.Context()
	address := c.Param("address")

	// 记录请求日志
	log.InfoWithContext(ctx, "收到获取地址余额请求", zap.String("address", address))

	// 参数验证
	if address == "" {
		c.JSON(http.StatusOK, gin.H{
			"status":  http.StatusBadRequest,
			"message": "地址参数不能为空",
		})
		return
	}

	// 调用业务逻辑层
	balanceData, err := s.addressLogic.GetAddressBalance(ctx, address)
	if err != nil {
		log.ErrorWithContext(ctx, "获取地址余额失败", zap.String("address", address), zap.Error(err))
		c.JSON(http.StatusOK, gin.H{
			"status":  http.StatusInternalServerError,
			"message": "获取地址余额失败: " + err.Error(),
		})
		return
	}

	// 返回成功响应
	c.JSON(http.StatusOK, gin.H{
		"status":  0,
		"address": address,
		"data":    balanceData,
	})
}

// GetAddressFrozenBalance 获取地址冻结余额
// @Router /v1/tbc/main/address/{address}/get/balance/frozen [get]
func (s *AddressService) GetAddressFrozenBalance(c *gin.Context) {
	// 获取上下文和参数
	ctx := c.Request.Context()
	address := c.Param("address")

	// 记录请求日志
	log.InfoWithContext(ctx, "收到获取地址冻结余额请求", zap.String("address", address))

	// 参数验证
	if address == "" {
		c.JSON(http.StatusOK, gin.H{
			"status":  http.StatusBadRequest,
			"message": "地址参数不能为空",
		})
		return
	}

	// 调用业务逻辑层
	frozenBalanceData, err := s.addressLogic.GetAddressFrozenBalance(ctx, address)
	if err != nil {
		log.ErrorWithContext(ctx, "获取地址冻结余额失败", zap.String("address", address), zap.Error(err))
		c.JSON(http.StatusOK, gin.H{
			"status":  http.StatusInternalServerError,
			"message": "获取地址冻结余额失败: " + err.Error(),
		})
		return
	}

	// 返回成功响应
	c.JSON(http.StatusOK, gin.H{
		"status":  0,
		"address": address,
		"data":    frozenBalanceData,
	})
}
