package ft

import (
	"net/http"
	"strings"

	"ginproject/entity/constant"
	"ginproject/entity/ft"
	"ginproject/entity/utility"
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
		c.JSON(http.StatusOK, utility.NewErrorResponse(constant.CodeInvalidParams, "无效的请求参数"))
		return
	}

	log.InfoWithContextf(ctx, "获取FT余额请求: %v", req)

	// 调用逻辑层处理业务
	response, err := s.ftLogic.GetFtBalance(ctx, &req)
	if err != nil {
		log.ErrorWithContextf(ctx, "处理FT余额查询失败: %v", err)
		c.JSON(http.StatusOK, utility.NewErrorResponse(constant.CodeServerError, "查询FT余额失败"))
		return
	}

	// 返回成功响应
	c.JSON(http.StatusOK, response)
}

// GetFtHistoryByAddress 根据地址和合约ID获取FT交易历史
// 路由: GET /v1/tbc/main/ft/history/address/:address/contract/:contract_id/page/:page/size/:size
func (s *FtService) GetFtHistoryByAddress(c *gin.Context) {
	ctx := c.Request.Context()

	// 绑定请求参数
	var req ft.FtHistoryRequest
	if err := c.ShouldBindUri(&req); err != nil {
		log.ErrorWithContextf(ctx, "绑定请求参数失败: %v", err)
		c.JSON(http.StatusOK, utility.NewErrorResponse(constant.CodeInvalidParams, "无效的请求参数"))
		return
	}

	log.InfoWithContextf(ctx, "获取FT交易历史请求: 地址=%s, 合约ID=%s, 页码=%d, 每页大小=%d",
		req.Address, req.ContractId, req.Page, req.Size)

	// 调用逻辑层处理业务
	response, err := s.ftLogic.GetFtHistory(ctx, &req)
	if err != nil {
		log.ErrorWithContextf(ctx, "处理FT交易历史查询失败: %v", err)
		c.JSON(http.StatusOK, utility.NewErrorResponse(constant.CodeServerError, "查询FT交易历史失败"))
		return
	}

	// 返回成功响应
	c.JSON(http.StatusOK, response)
}

// GetFtUtxoByAddress 根据地址和合约ID获取FT UTXO列表
// 路由: GET /v1/tbc/main/ft/utxo/address/:address/contract/:contract_id
func (s *FtService) GetFtUtxoByAddress(c *gin.Context) {
	ctx := c.Request.Context()

	// 绑定请求参数
	var req ft.FtUtxoAddressRequest
	if err := c.ShouldBindUri(&req); err != nil {
		log.ErrorWithContextf(ctx, "绑定请求参数失败: %v", err)
		c.JSON(http.StatusOK, utility.NewErrorResponse(constant.CodeInvalidParams, "无效的请求参数"))
		return
	}

	log.InfoWithContextf(ctx, "获取FT UTXO请求: %v", req)

	// 调用逻辑层处理业务
	response, err := s.ftLogic.GetFtUtxosByAddress(ctx, &req)
	if err != nil {
		log.ErrorWithContextf(ctx, "处理FT UTXO查询失败: %v", err)
		c.JSON(http.StatusOK, utility.NewErrorResponse(constant.CodeServerError, "查询FT UTXO失败"))
		return
	}

	// 返回成功响应
	c.JSON(http.StatusOK, response)
}

// GetFtInfoByContractId 根据合约ID获取FT信息
// 路由: GET /v1/tbc/main/ft/info/contract/id/:contract_id
func (s *FtService) GetFtInfoByContractId(c *gin.Context) {
	ctx := c.Request.Context()

	// 绑定请求参数
	var req ft.FtInfoContractIdRequest
	if err := c.ShouldBindUri(&req); err != nil {
		log.ErrorWithContextf(ctx, "绑定请求参数失败: %v", err)
		c.JSON(http.StatusOK, utility.NewErrorResponse(constant.CodeInvalidParams, "无效的请求参数"))
		return
	}

	log.InfoWithContextf(ctx, "获取FT信息请求: 合约ID=%s", req.ContractId)

	// 调用逻辑层处理业务
	response, err := s.ftLogic.GetFtInfoByContractId(ctx, &req)
	if err != nil {
		log.ErrorWithContextf(ctx, "处理FT信息查询失败: %v", err)
		c.JSON(http.StatusOK, utility.NewErrorResponse(constant.CodeServerError, "查询FT信息失败"))
		return
	}

	// 返回成功响应
	c.JSON(http.StatusOK, response)
}

// GetMultiFtBalanceByAddress 获取地址持有的多个代币余额
// 路由: POST /v1/tbc/main/ft/balance/address/:address/contract/ids
func (s *FtService) GetMultiFtBalanceByAddress(c *gin.Context) {
	ctx := c.Request.Context()

	// 从URI获取地址参数
	address := c.Param("address")
	if address == "" {
		log.ErrorWithContextf(ctx, "地址参数为空")
		c.JSON(http.StatusOK, utility.NewErrorResponse(constant.CodeInvalidParams, "地址不能为空"))
		return
	}

	// 绑定JSON请求体
	var reqBody struct {
		FtContractId []string `json:"ftContractId" binding:"required"`
	}

	if err := c.ShouldBindJSON(&reqBody); err != nil {
		log.ErrorWithContextf(ctx, "绑定JSON请求体失败: %v", err)
		c.JSON(http.StatusOK, utility.NewErrorResponse(constant.CodeInvalidParams, "无效的请求参数"))
		return
	}

	// 构造完整请求对象
	req := &ft.FtBalanceMultiContractRequest{
		Address:      address,
		FtContractId: reqBody.FtContractId,
	}

	log.InfoWithContextf(ctx, "获取多个FT余额请求: 地址=%s, 合约数量=%d", req.Address, len(req.FtContractId))

	// 调用逻辑层处理业务
	responseList, err := s.ftLogic.GetMultiFtBalanceByAddress(ctx, req)
	if err != nil {
		log.ErrorWithContextf(ctx, "处理多个FT余额查询失败: %v", err)
		c.JSON(http.StatusOK, utility.NewErrorResponse(constant.CodeServerError, "查询多个FT余额失败"))
		return
	}

	// 返回成功响应
	c.JSON(http.StatusOK, responseList)
}

// GetPoolNFTInfoByContractId 根据合约ID获取NFT池信息
// 路由: GET /v1/tbc/main/ft/pool/nft/info/contract/id/:ft_contract_id
func (s *FtService) GetPoolNFTInfoByContractId(c *gin.Context) {
	ctx := c.Request.Context()

	// 绑定请求参数
	var req ft.TBC20PoolNFTInfoRequest
	if err := c.ShouldBindUri(&req); err != nil {
		log.ErrorWithContextf(ctx, "绑定请求参数失败: %v", err)
		c.JSON(http.StatusOK, utility.NewErrorResponse(constant.CodeInvalidParams, "无效的请求参数"))
		return
	}

	log.InfoWithContextf(ctx, "获取NFT池信息请求: 合约ID=%s", req.FtContractId)

	// 调用逻辑层处理业务
	response, err := s.ftLogic.GetNFTPoolInfoByContractId(ctx, &req)
	if err != nil {
		log.ErrorWithContextf(ctx, "处理NFT池信息查询失败: %v", err)

		// 判断是否是未找到NFT池的错误
		if err.Error() == "No NFT pool info found." {
			// 返回特定的错误格式
			c.JSON(http.StatusOK, ft.ErrorResponse{
				Error: "No pool NFT found.",
			})
			return
		} else if strings.HasPrefix(err.Error(), "Decode pool NFT failed") {
			// 返回解码失败的错误
			c.JSON(http.StatusOK, ft.ErrorResponse{
				Error: "Decode pool NFT failed.",
			})
			return
		}

		// 其他服务器错误
		c.JSON(http.StatusOK, ft.ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	// 返回成功响应
	c.JSON(http.StatusOK, response)
}

// GetFtTokenList 获取代币列表
// 路由: GET /v1/tbc/main/ft/tokens/page/{page}/size/{size}/orderby/{order_by}
func (s *FtService) GetFtTokenList(c *gin.Context) {
	ctx := c.Request.Context()

	// 绑定请求参数
	var req ft.FtTokenListRequest
	if err := c.ShouldBindUri(&req); err != nil {
		log.ErrorWithContextf(ctx, "绑定请求参数失败: %v", err)
		c.JSON(http.StatusOK, utility.NewErrorResponse(constant.CodeInvalidParams, "无效的请求参数"))
		return
	}

	log.InfoWithContextf(ctx, "获取代币列表请求: 页码=%d, 每页大小=%d, 排序字段=%s",
		req.Page, req.Size, req.OrderBy)

	// 调用逻辑层处理业务
	response, err := s.ftLogic.GetFtTokenList(ctx, &req)
	if err != nil {
		log.ErrorWithContextf(ctx, "处理代币列表查询失败: %v", err)
		c.JSON(http.StatusOK, utility.NewErrorResponse(constant.CodeServerError, "查询代币列表失败"))
		return
	}

	// 返回成功响应
	c.JSON(http.StatusOK, response)
}
