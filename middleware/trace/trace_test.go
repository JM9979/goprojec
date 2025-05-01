package trace

import (
	"context"
	"testing"
)

func TestTraceIDGeneration(t *testing.T) {
	// 创建包含trace的context
	ctx := NewContext(context.Background(), "test-operation")

	// 提取trace ID和span ID
	traceID, spanID := ExtractIDs(ctx)

	// 验证trace ID和span ID不为空
	if traceID == "" {
		t.Error("TraceID should not be empty")
	}

	if spanID == "" {
		t.Error("SpanID should not be empty")
	}

	t.Logf("Generated TraceID: %s", traceID)
	t.Logf("Generated SpanID: %s", spanID)
}

func TestChildSpan(t *testing.T) {
	// 创建父context
	parentCtx := NewContext(context.Background(), "parent-operation")
	parentTraceID, parentSpanID := ExtractIDs(parentCtx)

	// 创建子span
	childCtx := WithNewSpan(parentCtx, "child-operation")
	childTraceID, childSpanID := ExtractIDs(childCtx)

	// 验证子span具有相同的trace ID但不同的span ID
	if childTraceID != parentTraceID {
		t.Errorf("Child span should have same trace ID. Parent: %s, Child: %s", parentTraceID, childTraceID)
	}

	if childSpanID == parentSpanID {
		t.Error("Child span should have different span ID")
	}

	t.Logf("Parent TraceID: %s, SpanID: %s", parentTraceID, parentSpanID)
	t.Logf("Child TraceID: %s, SpanID: %s", childTraceID, childSpanID)
}

func TestCurrentSpan(t *testing.T) {
	// 创建context
	ctx := NewContext(context.Background(), "test-current-span")

	// 获取当前span
	span := CurrentSpan(ctx)

	// 验证span不为nil
	if span == nil {
		t.Error("CurrentSpan should not return nil")
	}

	// 验证span context是有效的
	if !span.SpanContext().IsValid() {
		t.Error("Span context should be valid")
	}
}

func TestEndSpan(t *testing.T) {
	// 这个测试主要是确保EndSpan不会panic
	ctx := NewContext(context.Background(), "test-end-span")

	// 结束span不应该导致panic
	EndSpan(ctx)
}
