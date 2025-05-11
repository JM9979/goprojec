package chain_info_service

import (
	"net/http"

	"ginproject/entity/block"
	"ginproject/middleware/log"
	"ginproject/repo/rpc/blockchain"

	"github.com/gin-gonic/gin"
)

// ChainInfoService 区块链信息服务接口
type ChainInfoService interface {
	GetChainInfo(c *gin.Context)
}

// chainInfoService 区块链信息服务实现
type chainInfoService struct{}

// NewChainInfoService 创建区块链信息服务实例
func NewChainInfoService() ChainInfoService {
	return &chainInfoService{}
}

// GetChainInfo 获取区块链信息
func (s *chainInfoService) GetChainInfo(c *gin.Context) {
	ctx := c.Request.Context()
	log.InfoWithContext(ctx, "获取区块链信息")

	// 调用RPC获取区块链信息
	chainInfoChan := blockchain.FetchChainInfo(ctx)
	result := <-chainInfoChan
	if result.Error != nil {
		log.ErrorWithContext(ctx, "获取区块链信息失败", "error", result.Error)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取区块链信息失败"})
		return
	}

	chainInfoData, ok := result.Result.(map[string]interface{})
	if !ok {
		log.ErrorWithContext(ctx, "区块链信息格式不正确")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取区块链信息失败"})
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
	log.InfoWithContext(ctx, "成功获取区块链信息",
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
