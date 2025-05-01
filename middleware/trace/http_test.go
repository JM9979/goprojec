package trace

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHTTPMiddleware(t *testing.T) {
	// 创建一个测试处理程序
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 从请求上下文中提取trace信息
		traceID, spanID := ExtractIDs(r.Context())

		// 检查trace信息是否存在
		if traceID == "" {
			t.Error("TraceID should not be empty in request context")
		}
		if spanID == "" {
			t.Error("SpanID should not be empty in request context")
		}

		// 写入响应
		w.WriteHeader(http.StatusOK)
	})

	// 应用中间件
	handler := HTTPMiddleware(testHandler)

	// 创建测试请求
	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()

	// 处理请求
	handler.ServeHTTP(rr, req)

	// 检查响应
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// 检查响应头中是否包含trace信息
	if traceID := rr.Header().Get(TraceIDHeader); traceID == "" {
		t.Error("TraceID header should not be empty")
	}
	if spanID := rr.Header().Get(SpanIDHeader); spanID == "" {
		t.Error("SpanID header should not be empty")
	}
}

func TestHTTPMiddlewareWithExistingTrace(t *testing.T) {
	// 创建一个测试处理程序
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 从请求上下文中提取trace信息
		traceID, spanID := ExtractIDs(r.Context())

		// 检查trace信息是否存在
		if traceID == "" {
			t.Error("TraceID should not be empty in request context")
		}
		if spanID == "" {
			t.Error("SpanID should not be empty in request context")
		}

		// 写入响应
		w.WriteHeader(http.StatusOK)
	})

	// 应用中间件
	handler := HTTPMiddleware(testHandler)

	// 创建测试请求，并添加已有的trace ID
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set(TraceIDHeader, "test-trace-id")
	rr := httptest.NewRecorder()

	// 处理请求
	handler.ServeHTTP(rr, req)

	// 检查响应
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// 检查响应头中是否包含trace信息
	if traceID := rr.Header().Get(TraceIDHeader); traceID == "" {
		t.Error("TraceID header should not be empty")
	}
	if spanID := rr.Header().Get(SpanIDHeader); spanID == "" {
		t.Error("SpanID header should not be empty")
	}
}

func TestInjectTraceToRequest(t *testing.T) {
	// 创建一个包含trace的context
	ctx := NewContext(context.Background(), "test-operation")
	traceID, spanID := ExtractIDs(ctx)

	// 创建请求并设置context
	req := httptest.NewRequest("GET", "/test", nil)
	req = req.WithContext(ctx)

	// 注入trace信息到请求
	req = InjectTraceToRequest(req)

	// 检查请求头中是否包含trace信息
	if reqTraceID := req.Header.Get(TraceIDHeader); reqTraceID != traceID {
		t.Errorf("TraceID header mismatch: got %s want %s", reqTraceID, traceID)
	}
	if reqSpanID := req.Header.Get(SpanIDHeader); reqSpanID != spanID {
		t.Errorf("SpanID header mismatch: got %s want %s", reqSpanID, spanID)
	}
}

func TestHTTPClientMiddleware(t *testing.T) {
	// 创建测试服务器
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 检查请求头中是否包含trace信息
		if traceID := r.Header.Get(TraceIDHeader); traceID == "" {
			t.Error("TraceID header should not be empty")
		}
		if spanID := r.Header.Get(SpanIDHeader); spanID == "" {
			t.Error("SpanID header should not be empty")
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// 创建HTTP客户端并应用中间件
	client := &http.Client{}
	client = HTTPClientMiddleware(client)

	// 创建一个包含trace的context
	ctx := NewContext(context.Background(), "test-operation")

	// 创建请求
	req, _ := http.NewRequestWithContext(ctx, "GET", server.URL, nil)

	// 发送请求
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}
	defer resp.Body.Close()

	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Unexpected status code: got %d want %d", resp.StatusCode, http.StatusOK)
	}
}
