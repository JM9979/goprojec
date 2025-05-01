package trace

import (
	"github.com/gin-gonic/gin"
)

// GinMiddleware 返回一个Gin中间件，用于处理请求的trace
func GinMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检查请求头中是否已有trace ID
		var traceID, spanID string
		if existingTraceID := c.GetHeader(TraceIDHeader); existingTraceID != "" {
			// 如果请求头中已有trace ID，在现有的trace中创建新的span
			ctx := WithNewSpan(c.Request.Context(), c.Request.Method+" "+c.FullPath())
			c.Request = c.Request.WithContext(ctx)
		} else {
			// 如果没有trace ID，创建一个新的
			ctx := NewContext(c.Request.Context(), c.Request.Method+" "+c.FullPath())
			c.Request = c.Request.WithContext(ctx)
		}

		// 从context中提取trace ID和span ID
		traceID, spanID = ExtractIDs(c.Request.Context())

		// 添加到响应头
		c.Header(TraceIDHeader, traceID)
		c.Header(SpanIDHeader, spanID)

		// 在请求的上下文中存储trace ID和span ID，以便在处理函数中使用
		c.Set("TraceID", traceID)
		c.Set("SpanID", spanID)

		// 处理请求
		c.Next()

		// 结束span
		EndSpan(c.Request.Context())
	}
}

// GetTraceIDFromGin 从Gin上下文中获取TraceID
func GetTraceIDFromGin(c *gin.Context) string {
	if traceID, exists := c.Get("TraceID"); exists {
		return traceID.(string)
	}

	// 如果没有在Gin上下文中找到，尝试从请求上下文中获取
	traceID, _ := ExtractIDs(c.Request.Context())
	return traceID
}

// GetSpanIDFromGin 从Gin上下文中获取SpanID
func GetSpanIDFromGin(c *gin.Context) string {
	if spanID, exists := c.Get("SpanID"); exists {
		return spanID.(string)
	}

	// 如果没有在Gin上下文中找到，尝试从请求上下文中获取
	_, spanID := ExtractIDs(c.Request.Context())
	return spanID
}

// WithChildSpan 创建一个新的子span并返回更新后的Gin上下文
func WithChildSpan(c *gin.Context, operationName string) *gin.Context {
	// 在现有的trace中创建一个新的span
	ctx := WithNewSpan(c.Request.Context(), operationName)

	// 更新请求上下文
	c.Request = c.Request.WithContext(ctx)

	// 从新的上下文中提取trace ID和span ID
	traceID, spanID := ExtractIDs(ctx)

	// 更新Gin上下文中存储的trace ID和span ID
	c.Set("TraceID", traceID)
	c.Set("SpanID", spanID)

	return c
}
