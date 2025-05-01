package main

import (
	"ginproject/middleware/trace"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	// 创建Gin引擎
	r := gin.Default()

	// 应用trace中间件
	r.Use(trace.GinMiddleware())

	// 定义路由
	r.GET("/", func(c *gin.Context) {
		// 从Gin上下文中获取trace ID和span ID
		traceID := trace.GetTraceIDFromGin(c)
		spanID := trace.GetSpanIDFromGin(c)

		// 记录trace信息
		log.Printf("处理请求 - Trace ID: %s, Span ID: %s", traceID, spanID)

		// 创建子span进行数据库操作
		c = trace.WithChildSpan(c, "database-query")
		// 模拟数据库操作
		time.Sleep(100 * time.Millisecond)
		dbTraceID := trace.GetTraceIDFromGin(c)
		dbSpanID := trace.GetSpanIDFromGin(c)

		// 返回结果
		c.JSON(http.StatusOK, gin.H{
			"message":               "Hello, World!",
			"trace_id":              traceID,
			"span_id":               spanID,
			"db_operation_trace_id": dbTraceID,
			"db_operation_span_id":  dbSpanID,
		})
	})

	// 嵌套路由的例子
	r.GET("/nested", func(c *gin.Context) {
		// 获取当前span的trace ID
		traceID := trace.GetTraceIDFromGin(c)
		spanID := trace.GetSpanIDFromGin(c)
		log.Printf("外层处理 - Trace ID: %s, Span ID: %s", traceID, spanID)

		// 创建一个子操作
		handleNestedOperation(c)

		c.JSON(http.StatusOK, gin.H{
			"message":  "Nested operation completed",
			"trace_id": traceID,
			"span_id":  spanID,
		})
	})

	// 启动服务器
	log.Println("启动服务器在 http://localhost:8080")
	r.Run(":8080")
}

func handleNestedOperation(c *gin.Context) {
	// 创建子span
	c = trace.WithChildSpan(c, "nested-operation")

	// 获取子span的trace ID和span ID
	traceID := trace.GetTraceIDFromGin(c)
	spanID := trace.GetSpanIDFromGin(c)
	log.Printf("嵌套操作 - Trace ID: %s, Span ID: %s", traceID, spanID)

	// 模拟一些处理
	time.Sleep(50 * time.Millisecond)

	// 进一步嵌套操作
	handleDeepNestedOperation(c)
}

func handleDeepNestedOperation(c *gin.Context) {
	// 创建更深层次的子span
	c = trace.WithChildSpan(c, "deep-nested-operation")

	// 获取深层子span的trace ID和span ID
	traceID := trace.GetTraceIDFromGin(c)
	spanID := trace.GetSpanIDFromGin(c)
	log.Printf("深层嵌套操作 - Trace ID: %s, Span ID: %s", traceID, spanID)

	// 模拟一些处理
	time.Sleep(25 * time.Millisecond)
}
