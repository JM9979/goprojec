package ft

import (
	"net/http"

	"ginproject/entity/ft"
	ftlogic "ginproject/logic/ft"
	"ginproject/middleware/log"

	"github.com/gin-gonic/gin"
)

// FtService FT代币服务
type FtService struct {
	ftLogic *ftlogic.FtLogic
}

// NewFtService 创建FtService实例
func NewFtService() *FtService {
	return &FtService{
		ftLogic: ftlogic.NewFtLogic(),
	}
}

// GetFtBalanceByAddress 根据地址和合约ID获取FT余额
// 路由: GET /v1/tbc/main/ft/balance/address/:address/contract/:contract_id
func (s *FtService) GetFtBalanceByAddress(c *gin.Context) {
	ctx := c.Request.Context()

	// 绑定请求参数
	var req ft.FtBalanceAddressRequest
	if err := c.ShouldBindUri(&req); err != nil {
		log.ErrorWithContextf(ctx, "绑定请求参数失败: %v", err)
		c.JSON(http.StatusOK, ft.NewErrorResponse(ft.CodeInvalidParams, "无效的请求参数"))
		return
	}

	log.InfoWithContextf(ctx, "获取FT余额请求: %v", req)

	// 调用逻辑层处理业务
	response, err := s.ftLogic.GetFtBalance(ctx, &req)
	if err != nil {
		log.ErrorWithContextf(ctx, "处理FT余额查询失败: %v", err)
		c.JSON(http.StatusOK, ft.NewErrorResponse(ft.CodeServerError, "查询FT余额失败"))
		return
	}

	// 返回成功响应
	c.JSON(http.StatusOK, response)
}
