package main

import (
	"context"
	"fmt"
	"time"

	"ginproject/middleware/trace"
)

func main() {
	// 1. 创建根context，包含trace ID和span ID
	ctx := trace.NewContext(context.Background(), "main-operation")

	// 2. 从context中提取trace ID和span ID
	traceID, spanID := trace.ExtractIDs(ctx)
	fmt.Printf("主函数 - Trace ID: %s\n", traceID)
	fmt.Printf("主函数 - Span ID: %s\n", spanID)

	// 3. 使用包含trace信息的context调用其他函数
	processRequest(ctx)

	// 4. 最后结束span
	trace.EndSpan(ctx)
}

func processRequest(ctx context.Context) {
	// 1. 在同一个trace下创建新的span
	ctx = trace.WithNewSpan(ctx, "process-request")
	defer trace.EndSpan(ctx)

	// 2. 获取当前函数的trace ID和span ID
	traceID, spanID := trace.ExtractIDs(ctx)
	fmt.Printf("处理请求 - Trace ID: %s\n", traceID)
	fmt.Printf("处理请求 - Span ID: %s\n", spanID)

	// 3. 模拟一些处理时间
	time.Sleep(100 * time.Millisecond)

	// 4. 调用其他服务或函数，继续传递context
	callDatabase(ctx)
}

func callDatabase(ctx context.Context) {
	// 1. 在同一个trace下创建新的span
	ctx = trace.WithNewSpan(ctx, "database-query")
	defer trace.EndSpan(ctx)

	// 2. 获取当前函数的trace ID和span ID
	traceID, spanID := trace.ExtractIDs(ctx)
	fmt.Printf("数据库操作 - Trace ID: %s\n", traceID)
	fmt.Printf("数据库操作 - Span ID: %s\n", spanID)

	// 3. 模拟数据库查询
	time.Sleep(50 * time.Millisecond)
}
