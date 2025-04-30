package service

import (
	"crypto/rand"
	"ginproject/entity"
	"ginproject/logic"
	"ginproject/middleware/log"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

// 创建一个tracer用于创建span
var tracer = otel.Tracer("ginproject.service.hello")

type HelloService struct {
	logic *logic.HelloLogic
}

func NewHelloService() *HelloService {
	return &HelloService{
		logic: &logic.HelloLogic{},
	}
}

func (s *HelloService) HelloHandler(c *gin.Context) {
	// 手动生成有效的TraceID
	traceID := generateRandomTraceID()

	// 使用OpenTelemetry创建一个span
	ctx, span := tracer.Start(c.Request.Context(), "hello-handler")
	defer span.End()

	// 将新context设置到请求中
	c.Request = c.Request.WithContext(ctx)

	var req entity.UsernameRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.ErrorContextf(ctx, "请求参数绑定失败: %v", err)
		c.JSON(400, gin.H{
			"code":     400,
			"message":  "无效的请求参数",
			"trace_id": traceID,
		})
		return
	}

	log.InfoContextf(ctx, "请求参数: %v", req)
	response := s.logic.SayHello(req)

	// 在响应中添加traceID
	response.TraceID = traceID

	log.InfoContextf(ctx, "响应结果: %v", response)
	c.JSON(200, response)
}

// generateRandomTraceID 生成一个随机的有效TraceID
func generateRandomTraceID() string {
	var tid [16]byte
	_, err := rand.Read(tid[:])
	if err != nil {
		// 如果随机数生成失败，确保至少有一个非零字节
		tid[0] = 1
	}

	// 确保生成的ID是有效的（至少有一个非零字节）
	var allZeros = true
	for _, b := range tid {
		if b != 0 {
			allZeros = false
			break
		}
	}

	// 如果全为零，设置第一个字节为1
	if allZeros {
		tid[0] = 1
	}

	traceID := trace.TraceID(tid)
	return traceID.String()
}
