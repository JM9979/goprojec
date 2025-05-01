package service

import (
	"ginproject/entity"
	"ginproject/logic"
	"ginproject/middleware/log"
	"ginproject/middleware/trace"

	"github.com/gin-gonic/gin"
)

type HelloService struct {
	logic *logic.HelloLogic
}

func NewHelloService() *HelloService {
	return &HelloService{
		logic: &logic.HelloLogic{},
	}
}

func (s *HelloService) HelloHandler(c *gin.Context) {

	// 使用OpenTelemetry创建一个span
	ctx := trace.NewContext(c.Request.Context(), "hello-handler")
	defer trace.EndSpan(ctx)
	traceID, spanID := trace.ExtractIDs(ctx)
	// 将新context设置到请求中
	c.Request = c.Request.WithContext(ctx)

	var req entity.UsernameRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.ErrorWithContextf(ctx, "请求参数绑定失败: %v", err)
		c.JSON(400, gin.H{
			"code":     400,
			"message":  "无效的请求参数",
			"trace_id": traceID,
			"span_id":  spanID,
		})
		return
	}

	log.InfoWithContextf(ctx, "请求参数: %v", req)
	response := s.logic.SayHello(req)

	// 在响应中添加traceID
	response.TraceID = traceID

	log.InfoWithContextf(ctx, "响应结果: %v", response)
	c.JSON(200, response)
}

