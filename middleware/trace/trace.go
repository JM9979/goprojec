package trace

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.opentelemetry.io/otel/trace"
)

const (
	// TraceIDHeader HTTP请求头中存储trace ID的键
	TraceIDHeader = "X-Trace-ID"
	// SpanIDHeader HTTP请求头中存储span ID的键
	SpanIDHeader = "X-Span-ID"
)

var (
	// 全局tracer实例
	tracer trace.Tracer
	// 服务名称
	serviceName string
)

// InitTracer 使用服务名初始化追踪器
func InitTracer(name string) {
	serviceName = name

	// 创建资源配置
	res := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceNameKey.String(serviceName),
	)

	// 创建TracerProvider
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)

	// 设置全局TracerProvider
	otel.SetTracerProvider(tp)

	// 更新tracer
	tracer = otel.Tracer(serviceName + ".service")
}

// NewContext 创建一个包含新trace和span的context
func NewContext(ctx context.Context, operationName string) context.Context {
	ctx, _ = tracer.Start(ctx, operationName)
	return ctx
}

// WithNewSpan 在现有的trace中创建一个新的span
func WithNewSpan(ctx context.Context, operationName string) context.Context {
	ctx, _ = tracer.Start(ctx, operationName)
	return ctx
}

// WithAttributes 在现有的trace中创建一个带有属性的新span
func WithAttributes(ctx context.Context, operationName string, attrs ...attribute.KeyValue) context.Context {
	ctx, _ = tracer.Start(ctx, operationName, trace.WithAttributes(attrs...))
	return ctx
}

// WithSpanKind 在现有的trace中创建一个指定类型的新span
func WithSpanKind(ctx context.Context, operationName string, kind trace.SpanKind) context.Context {
	ctx, _ = tracer.Start(ctx, operationName, trace.WithSpanKind(kind))
	return ctx
}

// ExtractIDs 从context中提取trace ID和span ID
func ExtractIDs(ctx context.Context) (string, string) {
	span := trace.SpanFromContext(ctx)
	return span.SpanContext().TraceID().String(), span.SpanContext().SpanID().String()
}

// CurrentSpan 从context中获取当前span
func CurrentSpan(ctx context.Context) trace.Span {
	return trace.SpanFromContext(ctx)
}

// SetTracer 设置自定义的Tracer
func SetTracer(customTracer trace.Tracer) {
	if customTracer != nil {
		tracer = customTracer
	}
}

// EndSpan 结束context中的span
func EndSpan(ctx context.Context) {
	span := trace.SpanFromContext(ctx)
	span.End()
}
