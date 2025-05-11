package mempool_service

import (
	"ginproject/middleware/log"
	"ginproject/repo/rpc/blockchain"
	"net/http"

	"github.com/gin-gonic/gin"
)

// MempoolService 内存池服务接口
type MempoolService interface {
	GetMemPoolTxs(c *gin.Context)
}

// mempoolService 内存池服务实现
type mempoolService struct{}

// NewMempoolService 创建内存池服务实例
func NewMempoolService() MempoolService {
	return &mempoolService{}
}

// GetMemPoolTxs 获取内存池中的交易
func (s *mempoolService) GetMemPoolTxs(c *gin.Context) {
	ctx := c.Request.Context()
	log.InfoWithContext(ctx, "获取内存池交易列表")

	// 调用RPC获取内存池交易列表
	mempoolTxsChan := blockchain.FetchMemPoolTxs(ctx)
	result := <-mempoolTxsChan
	if result.Error != nil {
		log.ErrorWithContext(ctx, "获取内存池交易列表失败", "error", result.Error)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取内存池交易列表失败"})
		return
	}

	c.JSON(http.StatusOK, result.Result)
}
