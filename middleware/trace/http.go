package trace

import (
	"context"
	"net/http"
)

const (
	// TraceIDHeader HTTP请求头中存储trace ID的键
	TraceIDHeader = "X-Trace-ID"
	// SpanIDHeader HTTP请求头中存储span ID的键
	SpanIDHeader = "X-Span-ID"
)

// HTTPMiddleware 是一个HTTP中间件，用于处理传入请求的trace
func HTTPMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var ctx context.Context

		// 检查请求头中是否已有trace ID
		if traceID := r.Header.Get(TraceIDHeader); traceID != "" {
			// 如果已有trace ID，我们需要使用它创建一个新的span
			// 这里简单地创建一个新的context，实际上可能需要更复杂的逻辑来恢复trace
			ctx = r.Context()
			operationName := r.Method + " " + r.URL.Path
			ctx = WithNewSpan(ctx, operationName)
		} else {
			// 如果没有trace ID，创建一个新的
			operationName := r.Method + " " + r.URL.Path
			ctx = NewContext(r.Context(), operationName)
		}

		// 从context中提取trace ID和span ID
		traceID, spanID := ExtractIDs(ctx)

		// 添加到响应头
		w.Header().Set(TraceIDHeader, traceID)
		w.Header().Set(SpanIDHeader, spanID)

		// 使用新的context继续请求
		next.ServeHTTP(w, r.WithContext(ctx))

		// 结束span
		EndSpan(ctx)
	})
}

// ExtractTraceFromRequest 从HTTP请求中提取trace信息
func ExtractTraceFromRequest(r *http.Request) (traceID, spanID string) {
	ctx := r.Context()
	return ExtractIDs(ctx)
}

// InjectTraceToRequest 将trace信息注入到HTTP请求中
func InjectTraceToRequest(r *http.Request) *http.Request {
	ctx := r.Context()
	traceID, spanID := ExtractIDs(ctx)

	// 将trace ID和span ID添加到请求头
	r.Header.Set(TraceIDHeader, traceID)
	r.Header.Set(SpanIDHeader, spanID)

	return r
}

// HTTPClientMiddleware 是一个用于HTTP客户端的中间件，可以在发出请求时添加trace信息
func HTTPClientMiddleware(client *http.Client) *http.Client {
	transport := client.Transport
	if transport == nil {
		transport = http.DefaultTransport
	}

	client.Transport = &traceTransport{
		base: transport,
	}

	return client
}

type traceTransport struct {
	base http.RoundTripper
}

func (t *traceTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// 从context中提取trace信息
	ctx := req.Context()
	traceID, spanID := ExtractIDs(ctx)

	// 如果context中没有trace信息，创建一个新的
	if traceID == "" || spanID == "" {
		operationName := req.Method + " " + req.URL.Path
		ctx = NewContext(ctx, operationName)
		req = req.WithContext(ctx)
		traceID, spanID = ExtractIDs(ctx)
	}

	// 将trace信息添加到请求头
	req.Header.Set(TraceIDHeader, traceID)
	req.Header.Set(SpanIDHeader, spanID)

	// 执行请求
	resp, err := t.base.RoundTrip(req)

	// 结束span
	EndSpan(ctx)

	return resp, err
}
