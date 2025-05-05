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
		c.JSON(http.StatusBadRequest, gin.H{"error": "区块高度必须为整数"})
		return
	}

	// 验证参数
	if err := block.ValidateBlockHeight(height); err != nil {
		log.Error("区块高度验证失败", "height", height, "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 调用RPC获取区块信息
	blockData, err := blockchain.FetchBlockByHeight(c.Request.Context(), height)
	if err != nil {
		log.Error("获取区块数据失败", "height", height, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取区块数据失败"})
		return
	}

	c.JSON(http.StatusOK, blockData)
}

// GetBlockByHash 通过哈希获取区块详情
func (s *blockService) GetBlockByHash(c *gin.Context) {
	hash := c.Param("hash")

	// 验证参数
	if err := block.ValidateBlockHash(hash); err != nil {
		log.Error("区块哈希验证失败", "hash", hash, "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 调用RPC获取区块信息
	blockData, err := blockchain.FetchBlockByHash(c.Request.Context(), hash)
	if err != nil {
		log.Error("获取区块数据失败", "hash", hash, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取区块数据失败"})
		return
	}

	c.JSON(http.StatusOK, blockData)
}

// GetBlockHeaderByHeight 通过高度获取区块头信息
func (s *blockService) GetBlockHeaderByHeight(c *gin.Context) {
	heightStr := c.Param("height")
	height, err := strconv.ParseInt(heightStr, 10, 64)
	if err != nil {
		log.Error("解析区块高度失败", "height", heightStr, "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "区块高度必须为整数"})
		return
	}

	// 验证参数
	if err := block.ValidateBlockHeight(height); err != nil {
		log.Error("区块高度验证失败", "height", height, "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 调用RPC获取区块头信息
	headerData, err := blockchain.FetchBlockHeaderByHeight(c.Request.Context(), height)
	if err != nil {
		log.Error("获取区块头数据失败", "height", height, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取区块头数据失败"})
		return
	}

	c.JSON(http.StatusOK, headerData)
}

// GetBlockHeaderByHash 通过哈希获取区块头信息
func (s *blockService) GetBlockHeaderByHash(c *gin.Context) {
	hash := c.Param("hash")

	// 验证参数
	if err := block.ValidateBlockHash(hash); err != nil {
		log.Error("区块哈希验证失败", "hash", hash, "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 调用RPC获取区块头信息
	headerData, err := blockchain.FetchBlockHeaderByHash(c.Request.Context(), hash)
	if err != nil {
		log.Error("获取区块头数据失败", "hash", hash, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取区块头数据失败"})
		return
	}

	c.JSON(http.StatusOK, headerData)
}

// GetNearby10Headers 获取附近的10个区块头信息
func (s *blockService) GetNearby10Headers(c *gin.Context) {
	// 调用RPC获取最近10个区块头信息
	headersData, err := blockchain.FetchNearby10Headers(c.Request.Context())
	if err != nil {
		log.Error("获取最近10个区块头数据失败", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取最近10个区块头数据失败"})
		return
	}

	if len(headersData) == 0 {
		log.Warn("未找到区块头数据")
		c.JSON(http.StatusNotFound, gin.H{"error": "未找到区块头数据"})
		return
	}

	c.JSON(http.StatusOK, headersData)
}
