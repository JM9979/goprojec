package block_service

import (
	"ginproject/entity/block"
	"ginproject/middleware/log"
	"ginproject/repo/rpc/blockchain"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// BlockService 区块服务接口
type BlockService interface {
	GetBlockByHeight(c *gin.Context)
	GetBlockByHash(c *gin.Context)
	GetBlockHeaderByHeight(c *gin.Context)
	GetBlockHeaderByHash(c *gin.Context)
	GetNearby10Headers(c *gin.Context)
	GetChainInfo(c *gin.Context)
}

// blockService 区块服务实现
type blockService struct{}

// NewBlockService 创建区块服务实例
func NewBlockService() BlockService {
	return &blockService{}
}

// GetBlockByHeight 通过高度获取区块详情
func (s *blockService) GetBlockByHeight(c *gin.Context) {
	heightStr := c.Param("height")
	height, err := strconv.ParseInt(heightStr, 10, 64)
	if err != nil {
		log.Error("解析区块高度失败", "height", heightStr, "error", err)
		c.JSON(http.StatusOK, gin.H{"error": "区块高度必须为整数"})
		return
	}

	// 验证参数
	if err := block.ValidateBlockHeight(height); err != nil {
		log.Error("区块高度验证失败", "height", height, "error", err)
		c.JSON(http.StatusOK, gin.H{"error": err.Error()})
		return
	}

	// 调用RPC获取区块信息
	blockDataChan := blockchain.FetchBlockByHeight(c.Request.Context(), height)
	result := <-blockDataChan
	if result.Error != nil {
		log.Error("获取区块数据失败", "height", height, "error", result.Error)
		c.JSON(http.StatusOK, gin.H{"error": "获取区块数据失败"})
		return
	}

	c.JSON(http.StatusOK, result.Result)
}

// GetBlockByHash 通过哈希获取区块详情
func (s *blockService) GetBlockByHash(c *gin.Context) {
	hash := c.Param("hash")

	// 验证参数
	if err := block.ValidateBlockHash(hash); err != nil {
		log.Error("区块哈希验证失败", "hash", hash, "error", err)
		c.JSON(http.StatusOK, gin.H{"error": err.Error()})
		return
	}

	// 调用RPC获取区块信息
	blockDataChan := blockchain.FetchBlockByHash(c.Request.Context(), hash)
	result := <-blockDataChan
	if result.Error != nil {
		log.Error("获取区块数据失败", "hash", hash, "error", result.Error)
		c.JSON(http.StatusOK, gin.H{"error": "获取区块数据失败"})
		return
	}

	c.JSON(http.StatusOK, result.Result)
}

// GetBlockHeaderByHeight 通过高度获取区块头信息
func (s *blockService) GetBlockHeaderByHeight(c *gin.Context) {
	heightStr := c.Param("height")
	height, err := strconv.ParseInt(heightStr, 10, 64)
	if err != nil {
		log.Error("解析区块高度失败", "height", heightStr, "error", err)
		c.JSON(http.StatusOK, gin.H{"error": "区块高度必须为整数"})
		return
	}

	// 验证参数
	if err := block.ValidateBlockHeight(height); err != nil {
		log.Error("区块高度验证失败", "height", height, "error", err)
		c.JSON(http.StatusOK, gin.H{"error": err.Error()})
		return
	}

	// 调用RPC获取区块头信息
	headerDataChan := blockchain.FetchBlockHeaderByHeight(c.Request.Context(), height)
	result := <-headerDataChan
	if result.Error != nil {
		log.Error("获取区块头数据失败", "height", height, "error", result.Error)
		c.JSON(http.StatusOK, gin.H{"error": "获取区块头数据失败"})
		return
	}

	c.JSON(http.StatusOK, result.Result)
}

// GetBlockHeaderByHash 通过哈希获取区块头信息
func (s *blockService) GetBlockHeaderByHash(c *gin.Context) {
	hash := c.Param("hash")

	// 验证参数
	if err := block.ValidateBlockHash(hash); err != nil {
		log.Error("区块哈希验证失败", "hash", hash, "error", err)
		c.JSON(http.StatusOK, gin.H{"error": err.Error()})
		return
	}

	// 调用RPC获取区块头信息
	headerDataChan := blockchain.FetchBlockHeaderByHash(c.Request.Context(), hash)
	result := <-headerDataChan
	if result.Error != nil {
		log.Error("获取区块头数据失败", "hash", hash, "error", result.Error)
		c.JSON(http.StatusOK, gin.H{"error": "获取区块头数据失败"})
		return
	}

	c.JSON(http.StatusOK, result.Result)
}

// GetNearby10Headers 获取附近的10个区块头信息
func (s *blockService) GetNearby10Headers(c *gin.Context) {
	// 调用RPC获取最近10个区块头信息
	headersDataChan := blockchain.FetchNearby10Headers(c.Request.Context())
	result := <-headersDataChan
	if result.Error != nil {
		log.Error("获取最近10个区块头数据失败", "error", result.Error)
		c.JSON(http.StatusOK, gin.H{"error": "获取最近10个区块头数据失败"})
		return
	}

	headersData, ok := result.Result.([]interface{})
	if !ok || len(headersData) == 0 {
		log.Warn("未找到区块头数据")
		c.JSON(http.StatusNotFound, gin.H{"error": "未找到区块头数据"})
		return
	}

	c.JSON(http.StatusOK, headersData)
}

// GetChainInfo 获取区块链信息
func (s *blockService) GetChainInfo(c *gin.Context) {
	// 调用RPC获取区块链信息
	chainInfoChan := blockchain.FetchChainInfo(c.Request.Context())
	result := <-chainInfoChan
	if result.Error != nil {
		log.Error("获取区块链信息失败", "error", result.Error)
		c.JSON(http.StatusOK, gin.H{"error": "获取区块链信息失败"})
		return
	}

	chainInfoData, ok := result.Result.(map[string]interface{})
	if !ok {
		log.Error("区块链信息格式不正确")
		c.JSON(http.StatusOK, gin.H{"error": "获取区块链信息失败"})
		return
	}

	// 将map数据转换为ChainInfo结构
	chainInfo := &block.ChainInfo{
		BestBlockHash:        getString(chainInfoData, "bestblockhash"),
		Blocks:               getInt64(chainInfoData, "blocks"),
		Chain:                getString(chainInfoData, "chain"),
		ChainWork:            getString(chainInfoData, "chainwork"),
		Difficulty:           getFloat64(chainInfoData, "difficulty"),
		Headers:              getInt64(chainInfoData, "headers"),
		MedianTime:           getInt64(chainInfoData, "mediantime"),
		Pruned:               getBool(chainInfoData, "pruned"),
		VerificationProgress: getFloat64(chainInfoData, "verificationprogress"),
	}

	// 打印日志记录关键信息
	log.Info("成功获取区块链信息",
		"blocks", chainInfo.Blocks,
		"chain", chainInfo.Chain,
		"difficulty", chainInfo.Difficulty)

	c.JSON(http.StatusOK, chainInfo)
}

// 以下是辅助函数，用于安全地获取map中的各种类型值
func getString(data map[string]interface{}, key string) string {
	if val, ok := data[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

func getInt64(data map[string]interface{}, key string) int64 {
	if val, ok := data[key]; ok {
		switch v := val.(type) {
		case float64:
			return int64(v)
		case int64:
			return v
		case int:
			return int64(v)
		}
	}
	return 0
}

func getFloat64(data map[string]interface{}, key string) float64 {
	if val, ok := data[key]; ok {
		if f, ok := val.(float64); ok {
			return f
		}
	}
	return 0
}

func getBool(data map[string]interface{}, key string) bool {
	if val, ok := data[key]; ok {
		if b, ok := val.(bool); ok {
			return b
		}
	}
	return false
}
