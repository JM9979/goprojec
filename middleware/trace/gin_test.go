package trace

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestGinMiddleware(t *testing.T) {
	// 设置Gin为测试模式
	gin.SetMode(gin.TestMode)

	// 创建Gin引擎
	r := gin.New()

	// 应用中间件
	r.Use(GinMiddleware())

	// 添加测试路由
	r.GET("/test", func(c *gin.Context) {
		// 获取trace ID和span ID
		traceID := GetTraceIDFromGin(c)
		spanID := GetSpanIDFromGin(c)

		// 检查trace ID和span ID
		if traceID == "" {
			t.Error("TraceID should not be empty")
		}
		if spanID == "" {
			t.Error("SpanID should not be empty")
		}

		c.String(http.StatusOK, "OK")
	})

	// 创建测试请求
	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	// 处理请求
	r.ServeHTTP(w, req)

	// 检查响应状态
	if w.Code != http.StatusOK {
		t.Errorf("Response code should be %d, got %d", http.StatusOK, w.Code)
	}

	// 检查响应头中是否包含trace信息
	if traceID := w.Header().Get(TraceIDHeader); traceID == "" {
		t.Error("TraceID header should not be empty")
	}
	if spanID := w.Header().Get(SpanIDHeader); spanID == "" {
		t.Error("SpanID header should not be empty")
	}
}

func TestWithChildSpan(t *testing.T) {
	// 设置Gin为测试模式
	gin.SetMode(gin.TestMode)

	// 创建Gin引擎
	r := gin.New()

	// 应用中间件
	r.Use(GinMiddleware())

	// 添加测试路由
	r.GET("/test-child-span", func(c *gin.Context) {
		// 获取父span的trace ID和span ID
		parentTraceID := GetTraceIDFromGin(c)
		parentSpanID := GetSpanIDFromGin(c)

		// 创建子span
		c = WithChildSpan(c, "child-operation")

		// 获取子span的trace ID和span ID
		childTraceID := GetTraceIDFromGin(c)
		childSpanID := GetSpanIDFromGin(c)

		// 验证子span具有相同的trace ID
		if childTraceID != parentTraceID {
			t.Errorf("Child trace ID should match parent. Parent: %s, Child: %s", parentTraceID, childTraceID)
		}

		// 验证子span具有不同的span ID
		if childSpanID == parentSpanID {
			t.Error("Child span ID should be different from parent")
		}

		c.String(http.StatusOK, "OK")
	})

	// 创建测试请求
	req, _ := http.NewRequest("GET", "/test-child-span", nil)
	w := httptest.NewRecorder()

	// 处理请求
	r.ServeHTTP(w, req)

	// 检查响应状态
	if w.Code != http.StatusOK {
		t.Errorf("Response code should be %d, got %d", http.StatusOK, w.Code)
	}
}
