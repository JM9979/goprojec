package health

import (
	"net/http"

	"ginproject/middleware/log"

	"github.com/gin-gonic/gin"
)

// HealthService 健康检查服务
type HealthService struct {
}

// NewHealthService 创建HealthService实例
func NewHealthService() *HealthService {
	return &HealthService{}
}

// HealthCheck 健康检查
func (s *HealthService) HealthCheck(c *gin.Context) {
	ctx := c.Request.Context()
	log.InfoWithContext(ctx, "HealthCheck")
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "Turing API is running.",
		"data":    gin.H{},
	})
}
