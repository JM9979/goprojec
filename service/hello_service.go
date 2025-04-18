package service

import (
	"GinProject/entity"
	"GinProject/logic"

	"github.com/gin-gonic/gin"
	"golang.org/x/exp/slog"
)

type HelloService struct {
	logic *logic.HelloLogic
}

func NewHelloService() *HelloService {
	return &HelloService{
		logic: &logic.HelloLogic{},
	}
}

func (s *HelloService) RegisterRoutes(r *gin.Engine) {
	r.POST("/say/hello", s.helloHandler)
}

func (s *HelloService) helloHandler(c *gin.Context) {
	var req entity.UsernameRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		slog.ErrorCtx(c.Request.Context(), "请求参数绑定失败", "error", err)
		c.JSON(400, gin.H{
			"code":    400,
			"message": "无效的请求参数",
		})
		return
	}

	response := s.logic.SayHello(req)
	c.JSON(200, response)
}
