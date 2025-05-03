package main

import (
	"os"

	"ginproject/middleware/log"
	"ginproject/middleware/trace"
	"ginproject/repo"
	"ginproject/service"
	ft_service "ginproject/service/ft_service"
	tbcapi "ginproject/service/tbc_api"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

var router *gin.Engine

func init() {
	router = gin.New()
	router.Use(gin.Recovery())
	// 添加trace中间件
	router.Use(trace.GinMiddleware())
}

func main() {
	// 全局初始化
	if err := repo.Global_init(); err != nil {
		log.Error("全局初始化失败", zap.Error(err))
		os.Exit(1)
	}

	// 注册路由
	registerRoutes(router)

	// 创建HTTP服务器并启动
	srv := service.CreateServer(router)
	srv.Start()
}

func registerRoutes(r *gin.Engine) {
	// 创建API路由组，设置前缀
	apiGroup := r.Group("/v1/tbc/main")

	// 添加健康检查端点
	apiGroup.GET("/health", tbcapi.NewTbcApiService().HealthCheck)

	// 注册FT服务API
	ftService := ft_service.NewFtService()
	apiGroup.GET("/ft/balance/address/:address/contract/:contract_id", ftService.GetFtBalanceByAddress)
	apiGroup.GET("/ft/utxo/address/:address/contract/:contract_id", ftService.GetFtUtxoByAddress)
	apiGroup.GET("/ft/info/contract/id/:contract_id", ftService.GetFtInfoByContractId)
	apiGroup.POST("/ft/balance/address/:address/contract/ids", ftService.GetMultiFtBalanceByAddress)
	apiGroup.GET("/ft/pool/nft/info/contract/id/:ft_contract_id", ftService.GetPoolNFTInfoByContractId)
	apiGroup.GET("/ft/history/address/:address/contract/:contract_id/page/:page/size/:size", ftService.GetFtHistoryByAddress)
	// 添加获取代币列表的路由
	apiGroup.GET("/ft/tokens/page/:page/size/:size/orderby/:order_by", ftService.GetFtTokenList)
	// 添加解析FT交易历史的路由
	apiGroup.GET("/ft/decode/tx/history/:txid", ftService.DecodeFtTransactionHistory)
	// 添加获取代币相关流动池列表的路由
	apiGroup.GET("/ft/pools/of/token/contract/id/:ft_contract_id", ftService.GetPoolsOfTokenByContractId)
	// 添加获取代币历史交易记录的路由
	apiGroup.GET("/ft/token/history/contract/id/:ft_contract_id/page/:page/size/:size", ftService.GetTokenHistoryByContractId)
	// 添加获取池子历史记录的路由
	apiGroup.GET("/ft/pool/history/pool/id/:pool_id/page/:page/size/:size", ftService.GetPoolHistoryByPoolId)
	// 添加获取交易池列表的路由
	apiGroup.GET("/ft/pool/list/page/:page/size/:size", ftService.GetPoolList)
	// 添加获取地址持有的代币列表的路由
	apiGroup.GET("/ft/tokens/held/by/address/:address", ftService.GetTokenListHeldByAddress)
}
