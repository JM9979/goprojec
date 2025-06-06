package ft_service

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
		if err.Error() == "未找到NFT池" {
			// 返回特定的错误格式
			c.JSON(http.StatusOK, ft.ErrorResponse{
				Error: "No pool NFT found.",
			})
			return
		} else if strings.HasPrefix(err.Error(), "解码交易失败") {
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

// DecodeFtTransactionHistory 解析FT交易历史
// 路由: GET /v1/tbc/main/ft/decode/tx/history/:txid
func (s *FtService) DecodeFtTransactionHistory(c *gin.Context) {
	ctx := c.Request.Context()

	// 绑定请求参数
	var req ft.FtTxDecodeRequest
	if err := c.ShouldBindUri(&req); err != nil {
		log.ErrorWithContextf(ctx, "绑定请求参数失败: %v", err)
		c.JSON(http.StatusOK, utility.NewErrorResponse(constant.CodeInvalidParams, "无效的请求参数"))
		return
	}

	log.InfoWithContextf(ctx, "解析FT交易历史请求: 交易ID=%s", req.Txid)

	// 调用逻辑层处理业务
	response, err := s.ftLogic.DecodeFtTransactionHistory(ctx, &req)
	if err != nil {
		log.ErrorWithContextf(ctx, "处理FT交易解析失败: %v", err)
		c.JSON(http.StatusOK, utility.NewErrorResponse(constant.CodeServerError, "解析FT交易失败"))
		return
	}

	// 返回成功响应
	c.JSON(http.StatusOK, response)
}

// GetFtTokenListHeldByCombineScript 通过合并脚本获取代币列表
// 路由: GET /v1/tbc/main/ft/tokens/held/by/combine/script/:combine_script
func (s *FtService) GetFtTokenListHeldByCombineScript(c *gin.Context) {
	ctx := c.Request.Context()

	// 绑定请求参数
	var req ft.TBC20TokenListHeldByCombineScriptRequest
	if err := c.ShouldBindUri(&req); err != nil {
		log.ErrorWithContextf(ctx, "绑定请求参数失败: %v", err)
		c.JSON(http.StatusOK, utility.NewErrorResponse(constant.CodeInvalidParams, "无效的请求参数"))
		return
	}

	log.InfoWithContextf(ctx, "通过合并脚本获取代币列表请求: 合并脚本=%s", req.CombineScript)

	// 调用逻辑层处理业务
	response, err := s.ftLogic.GetTokensListHeldByCombineScript(ctx, &req)
	if err != nil {
		log.ErrorWithContextf(ctx, "处理通过合并脚本获取代币列表查询失败: %v", err)
		c.JSON(http.StatusInternalServerError, utility.NewErrorResponse(constant.CodeServerError, "查询代币列表失败"))
		return
	}

	// 返回成功响应
	c.JSON(http.StatusOK, response)
}

// GetTokenListHeldByAddress 获取地址持有的代币列表
// 路由: GET /v1/tbc/main/ft/tokens/held/by/address/:address
func (s *FtService) GetTokenListHeldByAddress(c *gin.Context) {
	ctx := c.Request.Context()

	// 绑定请求参数
	var req ft.TBC20TokenListHeldByAddressRequest
	if err := c.ShouldBindUri(&req); err != nil {
		log.ErrorWithContextf(ctx, "绑定请求参数失败: %v", err)
		c.JSON(http.StatusOK, utility.NewErrorResponse(constant.CodeInvalidParams, "无效的请求参数"))
		return
	}

	log.InfoWithContextf(ctx, "获取地址持有的代币列表请求: 地址=%s", req.Address)

	// 调用逻辑层处理业务
	response, err := s.ftLogic.GetTokensListHeldByAddress(ctx, &req)
	if err != nil {
		log.ErrorWithContextf(ctx, "处理地址持有的代币列表查询失败: %v", err)
		c.JSON(http.StatusOK, utility.NewErrorResponse(constant.CodeServerError, "查询地址持有的代币列表失败"))
		return
	}

	// 返回成功响应
	c.JSON(http.StatusOK, response)
}

// GetPoolsOfTokenByContractId 获取代币相关的流动池列表
// 路由: GET /v1/tbc/main/ft/pools/of/token/contract/id/:ft_contract_id
func (s *FtService) GetPoolsOfTokenByContractId(c *gin.Context) {
	ctx := c.Request.Context()

	// 绑定请求参数
	var req ft.TBC20PoolListRequest
	if err := c.ShouldBindUri(&req); err != nil {
		log.ErrorWithContextf(ctx, "绑定请求参数失败: %v", err)
		c.JSON(http.StatusOK, utility.NewErrorResponse(constant.CodeInvalidParams, "无效的请求参数"))
		return
	}

	log.InfoWithContextf(ctx, "获取代币相关流动池列表请求: 合约ID=%s", req.FtContractId)

	// 调用逻辑层处理业务
	response, err := s.ftLogic.GetPoolListByFtContractId(ctx, &req)
	if err != nil {
		log.ErrorWithContextf(ctx, "处理代币相关流动池列表查询失败: %v", err)
		c.JSON(http.StatusOK, utility.NewErrorResponse(constant.CodeServerError, "查询代币相关流动池列表失败"))
		return
	}

	// 返回成功响应
	c.JSON(http.StatusOK, response)
}

// GetTokenHistoryByContractId 根据合约ID获取代币交易历史
// 路由: GET /v1/tbc/main/ft/token/history/contract/id/:ft_contract_id/page/:page/size/:size
func (s *FtService) GetTokenHistoryByContractId(c *gin.Context) {
	ctx := c.Request.Context()

	// 绑定请求参数
	var req ft.FtTokenHistoryRequest
	if err := c.ShouldBindUri(&req); err != nil {
		log.ErrorWithContextf(ctx, "绑定请求参数失败: %v", err)
		c.JSON(http.StatusOK, utility.NewErrorResponse(constant.CodeInvalidParams, "无效的请求参数"))
		return
	}

	log.InfoWithContextf(ctx, "获取代币历史交易记录请求: 合约ID=%s, 页码=%d, 每页大小=%d",
		req.FtContractId, req.Page, req.Size)

	// 调用逻辑层处理业务
	response, err := s.ftLogic.GetTokenHistory(ctx, &req)
	if err != nil {
		log.ErrorWithContextf(ctx, "处理代币历史交易记录查询失败: %v", err)
		c.JSON(http.StatusOK, utility.NewErrorResponse(constant.CodeServerError, "查询代币历史交易记录失败"))
		return
	}

	// 返回成功响应
	c.JSON(http.StatusOK, response)
}

// GetPoolHistoryByPoolId 获取池子历史记录
// 路由: GET /v1/tbc/main/ft/pool/history/pool/id/{pool_id}/page/{page}/size/{size}
func (s *FtService) GetPoolHistoryByPoolId(c *gin.Context) {
	ctx := c.Request.Context()

	// 绑定请求参数
	var req ft.TBC20PoolHistoryRequest
	if err := c.ShouldBindUri(&req); err != nil {
		log.ErrorWithContextf(ctx, "绑定请求参数失败: %v", err)
		c.JSON(http.StatusOK, utility.NewErrorResponse(constant.CodeInvalidParams, "无效的请求参数"))
		return
	}

	log.InfoWithContextf(ctx, "获取池子历史记录请求: 池子ID=%s, 页码=%d, 每页大小=%d",
		req.PoolId, req.Page, req.Size)

	// 调用逻辑层处理业务
	response, err := s.ftLogic.GetPoolHistory(ctx, &req)
	if err != nil {
		log.ErrorWithContextf(ctx, "处理池子历史记录查询失败: %v", err)
		c.JSON(http.StatusOK, utility.NewErrorResponse(constant.CodeServerError, "查询池子历史记录失败"))
		return
	}

	// 返回成功响应
	c.JSON(http.StatusOK, response)
}

// GetPoolList 获取交易池列表
// 路由: GET /v1/tbc/main/ft/pool/list/page/{page}/size/{size}
func (s *FtService) GetPoolList(c *gin.Context) {
	ctx := c.Request.Context()

	// 绑定请求参数
	var req ft.TBC20PoolPageRequest
	if err := c.ShouldBindUri(&req); err != nil {
		log.ErrorWithContextf(ctx, "绑定请求参数失败: %v", err)
		c.JSON(http.StatusOK, utility.NewErrorResponse(constant.CodeInvalidParams, "无效的请求参数"))
		return
	}

	log.InfoWithContextf(ctx, "获取交易池列表请求: 页码=%d, 每页大小=%d", req.Page, req.Size)

	// 调用逻辑层处理业务
	response, err := s.ftLogic.GetAllPoolList(ctx, &req)
	if err != nil {
		log.ErrorWithContextf(ctx, "处理交易池列表查询失败: %v", err)
		c.JSON(http.StatusOK, utility.NewErrorResponse(constant.CodeServerError, "查询交易池列表失败"))
		return
	}

	// 返回成功响应
	c.JSON(http.StatusOK, response)
}

// GetHolderRankByContractId 获取代币持有者排名
// 路由: GET /v1/tbc/main/ft/holder/rank/contract/:contract_id/page/:page/size/:size
func (s *FtService) GetHolderRankByContractId(c *gin.Context) {
	ctx := c.Request.Context()

	// 绑定请求参数
	var req ft.FtHolderRankRequest
	if err := c.ShouldBindUri(&req); err != nil {
		log.ErrorWithContextf(ctx, "绑定请求参数失败: %v", err)
		c.JSON(http.StatusOK, utility.NewErrorResponse(constant.CodeInvalidParams, "无效的请求参数"))
		return
	}

	log.InfoWithContextf(ctx, "获取代币持有者排名请求: 合约ID=%s, 页码=%d, 每页大小=%d",
		req.ContractId, req.Page, req.Size)

	// 调用逻辑层处理业务
	response, err := s.ftLogic.GetFtHolderRank(ctx, &req)
	if err != nil {
		log.ErrorWithContextf(ctx, "处理代币持有者排名查询失败: %v", err)
		c.JSON(http.StatusOK, utility.NewErrorResponse(constant.CodeServerError, "查询代币持有者排名失败"))
		return
	}

	// 返回成功响应
	c.JSON(http.StatusOK, response)
}

// GetFtUtxoByCombineScript 根据合并脚本和合约ID获取FT UTXO列表
// 路由: GET /v1/tbc/main/ft/utxo/combine/script/:combine_script/contract/:contract_id
func (s *FtService) GetFtUtxoByCombineScript(c *gin.Context) {
	ctx := c.Request.Context()

	// 绑定请求参数
	var req ft.FtUtxoCombineScriptRequest
	if err := c.ShouldBindUri(&req); err != nil {
		log.ErrorWithContextf(ctx, "绑定请求参数失败: %v", err)
		c.JSON(http.StatusOK, utility.NewErrorResponse(constant.CodeInvalidParams, "无效的请求参数"))
		return
	}

	// 验证请求参数
	if err := req.Validate(); err != nil {
		log.ErrorWithContextf(ctx, "验证请求参数失败: %v", err)
		c.JSON(http.StatusOK, utility.NewErrorResponse(constant.CodeInvalidParams, err.Error()))
		return
	}

	log.InfoWithContextf(ctx, "获取基于合并脚本的FT UTXO请求: 合并脚本=%s, 合约ID=%s",
		req.CombineScript, req.ContractId)

	// 调用逻辑层处理业务
	response, err := s.ftLogic.GetFtUtxosByCombineScript(ctx, &req)
	if err != nil {
		log.ErrorWithContextf(ctx, "处理基于合并脚本的FT UTXO查询失败: %v", err)
		c.JSON(http.StatusOK, utility.NewErrorResponse(constant.CodeServerError, "查询FT UTXO失败"))
		return
	}

	// 返回成功响应
	c.JSON(http.StatusOK, response)
}

// GetFtBalanceByCombineScript 根据合并脚本和合约哈希获取FT余额
// 路由: GET /v1/tbc/main/ft/balance/combine/script/:combine_script/contract/:contract_hash
func (s *FtService) GetFtBalanceByCombineScript(c *gin.Context) {
	ctx := c.Request.Context()

	// 绑定请求参数
	var req ft.FtBalanceCombineScriptRequest
	if err := c.ShouldBindUri(&req); err != nil {
		log.ErrorWithContextf(ctx, "绑定请求参数失败: %v", err)
		c.JSON(http.StatusOK, utility.NewErrorResponse(constant.CodeInvalidParams, "无效的请求参数"))
		return
	}

	// 验证请求参数
	if err := req.Validate(); err != nil {
		log.ErrorWithContextf(ctx, "验证请求参数失败: %v", err)
		c.JSON(http.StatusOK, utility.NewErrorResponse(constant.CodeInvalidParams, err.Error()))
		return
	}

	log.InfoWithContextf(ctx, "获取基于合并脚本的FT余额请求: 合并脚本=%s, 合约哈希=%s",
		req.CombineScript, req.ContractHash)

	// 调用逻辑层处理业务
	response, err := s.ftLogic.GetFtBalanceByCombineScript(ctx, &req)
	if err != nil {
		log.ErrorWithContextf(ctx, "处理基于脚本的FT余额查询失败: %v", err)
		c.JSON(http.StatusOK, utility.NewErrorResponse(constant.CodeServerError, "查询FT余额失败"))
		return
	}

	// 返回成功响应
	c.JSON(http.StatusOK, response)
}

// GetLPUnspentByScriptHash 根据脚本哈希获取LP未花费交易输出
// 路由: GET /v1/tbc/main/ft/lp/unspent/by/script/hash/:script_hash
func (s *FtService) GetLPUnspentByScriptHash(c *gin.Context) {
	ctx := c.Request.Context()

	// 绑定请求参数
	var req ft.LPUnspentByScriptHashRequest
	if err := c.ShouldBindUri(&req); err != nil {
		log.ErrorWithContextf(ctx, "绑定请求参数失败: %v", err)
		c.JSON(http.StatusOK, utility.NewErrorResponse(constant.CodeInvalidParams, "无效的请求参数"))
		return
	}

	// 验证参数
	if err := req.Validate(); err != nil {
		log.ErrorWithContextf(ctx, "参数验证失败: %v", err)
		c.JSON(http.StatusOK, utility.NewErrorResponse(constant.CodeInvalidParams, err.Error()))
		return
	}

	log.InfoWithContextf(ctx, "获取LP未花费交易输出请求: 脚本哈希=%s", req.ScriptHash)

	// 调用逻辑层处理业务
	response, err := s.ftLogic.GetLPUnspentByScriptHash(ctx, &req)
	if err != nil {
		log.ErrorWithContextf(ctx, "处理LP未花费交易输出查询失败: %v", err)
		c.JSON(http.StatusOK, utility.NewErrorResponse(constant.CodeServerError, "查询LP未花费交易输出失败"))
		return
	}

	// 返回成功响应
	c.JSON(http.StatusOK, response)
}
