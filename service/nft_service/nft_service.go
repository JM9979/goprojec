package nft_service

import (
	"net/http"
	"strconv"

	"ginproject/entity/nft"
	nftLogic "ginproject/logic/nft"
	"ginproject/middleware/log"

	"github.com/gin-gonic/gin"
)

// NftService NFT服务结构体
type NftService struct {
	logic *nftLogic.NFTLogic
}

// NewNftService 创建新的NFT服务实例
func NewNftService() *NftService {
	return &NftService{
		logic: nftLogic.NewNFTLogic(),
	}
}

// GetNftsByContractIds 根据合约ID获取NFT信息
func (s *NftService) GetNftsByContractIds(c *gin.Context) {
	var req nft.NftsByContractIdsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数解析失败: " + err.Error()})
		return
	}

	// 参数校验
	if len(req.ContractList) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "合约ID列表不能为空"})
		return
	}

	// 调用API逻辑层
	response, err := s.logic.GetNftsByContractIds(c, req.ContractList, req.IfIconNeeded)
	if err != nil {
		log.ErrorWithContext(c, "获取合约ID列表NFT信息失败", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取合约ID列表NFT信息失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetCollectionsByAddress 获取地址的NFT集合
func (s *NftService) GetCollectionsByAddress(c *gin.Context) {
	// 获取路径参数
	address := c.Param("address")
	pageStr := c.Param("page")
	sizeStr := c.Param("size")

	// 参数转换
	page, err := strconv.Atoi(pageStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "页码必须为数字"})
		return
	}
	size, err := strconv.Atoi(sizeStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "每页数量必须为数字"})
		return
	}

	// 参数校验
	if address == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "地址不能为空"})
		return
	}
	if page < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "页码不能为负数"})
		return
	}
	if size <= 0 || size > 10000 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "每页记录数必须在1-10000之间"})
		return
	}

	// 调用API逻辑层
	response, err := s.logic.GetCollectionByAddressPageSize(c, address, page, size)
	if err != nil {
		log.ErrorWithContext(c, "获取地址NFT集合失败", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取地址NFT集合失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetNftsByAddress 获取地址持有的NFT资产
func (s *NftService) GetNftsByAddress(c *gin.Context) {
	// 获取路径参数
	address := c.Param("address")
	pageStr := c.Param("page")
	sizeStr := c.Param("size")

	// 获取查询参数
	ifExtraCollectionInfoStr := c.Query("if_extra_collection_info_needed")
	ifExtraCollectionInfo := false
	if ifExtraCollectionInfoStr == "true" {
		ifExtraCollectionInfo = true
	}

	// 参数转换
	page, err := strconv.Atoi(pageStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "页码必须为数字"})
		return
	}
	size, err := strconv.Atoi(sizeStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "每页数量必须为数字"})
		return
	}

	// 参数校验
	if address == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "地址不能为空"})
		return
	}
	if page < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "页码不能为负数"})
		return
	}
	if size <= 0 || size > 10000 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "每页记录数必须在1-10000之间"})
		return
	}

	// 调用API逻辑层
	response, err := s.logic.GetNftByAddressPageSize(c, address, page, size, ifExtraCollectionInfo)
	if err != nil {
		log.ErrorWithContext(c, "获取地址NFT资产失败", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取地址NFT资产失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetNftsByScriptHash 获取脚本哈希持有的NFT资产
func (s *NftService) GetNftsByScriptHash(c *gin.Context) {
	// 获取路径参数
	scriptHash := c.Param("script_hash")
	pageStr := c.Param("page")
	sizeStr := c.Param("size")

	// 参数转换
	page, err := strconv.Atoi(pageStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "页码必须为数字"})
		return
	}
	size, err := strconv.Atoi(sizeStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "每页数量必须为数字"})
		return
	}

	// 参数校验
	if scriptHash == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "脚本哈希不能为空"})
		return
	}
	if page < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "页码不能为负数"})
		return
	}
	if size <= 0 || size > 10000 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "每页记录数必须在1-10000之间"})
		return
	}

	log.InfoWithContext(c, "获取脚本哈希NFT资产", "scriptHash", scriptHash, "page", page, "size", size)
	// 调用API逻辑层
	response, err := s.logic.GetNftByScriptHashPageSize(c, scriptHash, page, size)
	if err != nil {
		log.ErrorWithContext(c, "获取脚本哈希NFT资产失败", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取脚本哈希NFT资产失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetNftsByCollectionId 获取集合的NFT资产
func (s *NftService) GetNftsByCollectionId(c *gin.Context) {
	// 获取路径参数
	collectionId := c.Param("collection_id")
	pageStr := c.Param("page")
	sizeStr := c.Param("size")

	// 参数转换
	page, err := strconv.Atoi(pageStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "页码必须为数字"})
		return
	}
	size, err := strconv.Atoi(sizeStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "每页数量必须为数字"})
		return
	}

	// 参数校验
	if collectionId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "集合ID不能为空"})
		return
	}
	if page < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "页码不能为负数"})
		return
	}
	if size <= 0 || size > 10000 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "每页记录数必须在1-10000之间"})
		return
	}

	// 调用API逻辑层
	response, err := s.logic.GetNftByCollectionIdPageSize(c, collectionId, page, size)
	if err != nil {
		log.ErrorWithContext(c, "获取集合NFT资产失败", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取集合NFT资产失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetNftHistory 获取地址的NFT交易历史
func (s *NftService) GetNftHistory(c *gin.Context) {
	// 获取路径参数
	address := c.Param("address")
	pageStr := c.Param("page")
	sizeStr := c.Param("size")

	// 参数转换
	page, err := strconv.Atoi(pageStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "页码必须为数字"})
		return
	}
	size, err := strconv.Atoi(sizeStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "每页数量必须为数字"})
		return
	}

	// 参数校验
	if address == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "地址不能为空"})
		return
	}
	if page < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "页码不能为负数"})
		return
	}
	if size <= 0 || size > 10000 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "每页记录数必须在1-10000之间"})
		return
	}

	// 调用API逻辑层
	response, err := s.logic.GetNftHistoryByAddress(c, address, page, size)
	if err != nil {
		log.ErrorWithContext(c, "获取NFT历史记录失败", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取NFT历史记录失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetAllCollections 获取所有NFT集合
func (s *NftService) GetAllCollections(c *gin.Context) {
	// 获取路径参数
	pageStr := c.Param("page")
	sizeStr := c.Param("size")

	// 参数转换
	page, err := strconv.Atoi(pageStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "页码必须为数字"})
		return
	}
	size, err := strconv.Atoi(sizeStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "每页数量必须为数字"})
		return
	}

	// 参数校验
	if page < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "页码不能为负数"})
		return
	}
	if size <= 0 || size > 10000 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "每页记录数必须在1-10000之间"})
		return
	}

	// 调用API逻辑层
	response, err := s.logic.GetCollectionsByPageSize(c, page, size)
	if err != nil {
		log.ErrorWithContext(c, "获取所有NFT集合失败", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取所有NFT集合失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetDetailCollectionInfo 获取集合详细信息
func (s *NftService) GetDetailCollectionInfo(c *gin.Context) {
	// 获取路径参数
	collectionId := c.Param("collection_id")

	// 参数校验
	if collectionId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "集合ID不能为空"})
		return
	}

	// 调用API逻辑层
	response, err := s.logic.GetDetailCollectionInfo(c, collectionId)
	if err != nil {
		log.ErrorWithContext(c, "获取集合详细信息失败", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取集合详细信息失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}
