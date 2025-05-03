package addressservice

import (
	"net/http"

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
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    http.StatusBadRequest,
			"message": "地址参数不能为空",
		})
		return
	}

	// 调用业务逻辑层
	utxos, err := s.addressLogic.GetAddressUnspentUtxos(ctx, address)
	if err != nil {
		log.ErrorWithContext(ctx, "获取地址UTXO失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    http.StatusInternalServerError,
			"message": "获取地址UTXO失败: " + err.Error(),
		})
		return
	}

	// 返回成功响应
	c.JSON(http.StatusOK, utxos)
}
