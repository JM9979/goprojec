package main

import (
	"os"

	"ginproject/middleware/log"
	"ginproject/middleware/trace"
	"ginproject/repo"
	"ginproject/service"
	address_service "ginproject/service/address_service"
	block_service "ginproject/service/block_service"
	chain_info_service "ginproject/service/chain_info_service"
	exchange_service "ginproject/service/exchange_service"
	ft_service "ginproject/service/ft_service"
	health_service "ginproject/service/health_service"
	mempool_service "ginproject/service/mempool_service"
	multisig_service "ginproject/service/multisig_service"
	nft_service "ginproject/service/nft_service"
	script_service "ginproject/service/script_service"
	transaction_service "ginproject/service/transaction"
	tx_broadcast_service "ginproject/service/tx_broadcast_service"

	"github.com/gin-gonic/gin"
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
		log.Error("全局初始化失败", "错误:", err)
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
	apiGroup.GET("/health", health_service.NewHealthService().HealthCheck)

	// 注册交易所服务API
	exchangeService := exchange_service.NewExchangeService()
	apiGroup.GET("/exchangerate", exchangeService.GetExchangeRate)

	// 注册FT服务API
	ftService := ft_service.NewFtService()
	apiGroup.GET("/ft/balance/address/:address/contract/:contract_id", ftService.GetFtBalanceByAddress)
	apiGroup.GET("/ft/utxo/address/:address/contract/:contract_id", ftService.GetFtUtxoByAddress)
	apiGroup.GET("/ft/info/contract/id/:contract_id", ftService.GetFtInfoByContractId)
	apiGroup.POST("/ft/balance/address/:address/contract/ids", ftService.GetMultiFtBalanceByAddress)
	apiGroup.GET("/ft/pool/nft/info/contract/id/:ft_contract_id", ftService.GetPoolNFTInfoByContractId)
	apiGroup.GET("/ft/lp/unspent/by/script/hash:script_hash", ftService.GetLPUnspentByScriptHash)
	apiGroup.GET("/ft/history/address/:address/contract/:contract_id/page/:page/size/:size", ftService.GetFtHistoryByAddress)
	// 添加获取代币列表的路由
	apiGroup.GET("/ft/tokens/page/:page/size/:size/orderby/:order_by", ftService.GetFtTokenList)
	// 添加通过合并脚本获取代币列表的路由
	apiGroup.GET("/ft/tokens/held/by/combine/script/:combine_script", ftService.GetFtTokenListHeldByCombineScript)
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
	// 添加获取代币持有者排名的路由
	apiGroup.GET("/ft/holder/rank/contract/:contract_id/page/:page/size/:size", ftService.GetHolderRankByContractId)
	// 添加根据合并脚本和合约ID获取FT UTXO的路由
	apiGroup.GET("/ft/utxo/combine/script/:combine_script/contract/:contract_id", ftService.GetFtUtxoByCombineScript)
	// 添加合并脚本和合约哈希获取FT余额的路由
	apiGroup.GET("/ft/balance/combine/script/:combine_script/contract/:contract_hash", ftService.GetFtBalanceByCombineScript)

	// 注册地址服务API
	addressService := address_service.NewAddressService()
	apiGroup.GET("/address/:address/unspent", addressService.GetAddressUnspentUtxos)
	// 添加获取地址历史交易的路由
	apiGroup.GET("/address/:address/history", addressService.GetAddressHistory)
	// 添加获取地址历史交易分页的路由
	apiGroup.GET("/address/:address/history/page/:page", addressService.GetAddressHistoryPaged)
	// 添加获取地址余额的路由
	apiGroup.GET("/address/:address/get/balance", addressService.GetAddressBalance)
	// 添加获取地址冻结余额的路由
	apiGroup.GET("/address/:address/get/balance/frozen", addressService.GetAddressFrozenBalance)

	// 注册区块服务API
	blockService := block_service.NewBlockService()
	// 添加通过高度获取区块详情的路由
	apiGroup.GET("/block/height/:height", blockService.GetBlockByHeight)
	// 添加通过哈希获取区块详情的路由
	apiGroup.GET("/block/hash/:hash", blockService.GetBlockByHash)
	// 添加通过高度获取区块头信息的路由
	apiGroup.GET("/block/height/:height/header", blockService.GetBlockHeaderByHeight)
	// 添加通过哈希获取区块头信息的路由
	apiGroup.GET("/block/hash/:hash/header", blockService.GetBlockHeaderByHash)
	// 添加获取附近10个区块头信息的路由
	apiGroup.GET("/block/headers", blockService.GetNearby10Headers)

	// 注册区块链信息服务API
	chainInfoService := chain_info_service.NewChainInfoService()
	// 添加获取区块链信息的路由
	apiGroup.GET("/chain/info", chainInfoService.GetChainInfo)

	// 注册内存池服务API
	mempoolService := mempool_service.NewMempoolService()
	// 添加获取内存池交易列表的路由
	apiGroup.GET("/mempool/mempool/txs", mempoolService.GetMemPoolTxs)

	// 注册脚本服务API
	scriptService := script_service.NewScriptService()
	apiGroup.GET("/script/hash/:script_hash/unspent", scriptService.GetScriptUnspent)
	apiGroup.GET("/script/hash/:script_hash/history", scriptService.GetScriptHistory)

	// 注册多签名服务API
	multisigService := multisig_service.NewMultisigService()
	// 添加根据地址获取多签名地址及其公钥列表的路由
	apiGroup.GET("/multisig/pubkeys/address/:address", multisigService.GetMultiWalletByAddress)

	// 注册NFT服务API
	nftService := nft_service.NewNftService()
	// 1. 获取地址的NFT集合
	apiGroup.GET("/nft/collection/address/:address/page/:page/size/:size", nftService.GetCollectionsByAddress)
	// 2. 获取地址的NFT资产
	apiGroup.GET("/nft/address/:address/page/:page/size/:size", nftService.GetNftsByAddress)
	// 3. 获取脚本哈希的NFT资产
	apiGroup.GET("/nft/script/hash/:script_hash/page/:page/size/:size", nftService.GetNftsByScriptHash)
	// 4. 获取集合的NFT资产
	apiGroup.GET("/nft/collection/id/:collection_id/page/:page/size/:size", nftService.GetNftsByCollectionId)
	// 5. 获取地址的NFT交易历史
	apiGroup.GET("/nft/history/address/:address/page/:page/size/:size", nftService.GetNftHistory)
	// 6. 获取所有NFT集合
	apiGroup.GET("/nft/collections/page/:page/size/:size", nftService.GetAllCollections)
	// 7. 获取集合详细信息
	apiGroup.GET("/nft/collection/info/:collection_id", nftService.GetDetailCollectionInfo)
	// 8. 根据合约ID获取NFT信息
	apiGroup.POST("/nft/infos/contract_ids", nftService.GetNftsByContractIds)

	// 注册交易广播服务API
	txBroadcastService := tx_broadcast_service.NewTxBroadcastService()
	// 广播单笔原始交易
	apiGroup.POST("/broadcast/tx/raw", txBroadcastService.BroadcastTxRaw)
	// 批量广播原始交易
	apiGroup.POST("/broadcast/txs/raw", txBroadcastService.BroadcastTxsRaw)

	// 注册交易服务API
	txService := transaction_service.NewTransactionService()
	// 广播单笔原始交易
	apiGroup.POST("/tx/raw", txService.BroadcastTxRaw)
	// 解码原始交易
	apiGroup.POST("/tx/raw/decode", txService.DecodeTxRaw)
	// 获取交易原始十六进制数据
	apiGroup.GET("/tx/hex/:txid", txService.GetTxRawHex)
	// 通过交易ID解码交易
	apiGroup.GET("/tx/hex/:txid/decode", txService.DecodeTxByHash)
	// 获取交易输入数据
	apiGroup.POST("/tx/vins", txService.GetTxVins)
}
