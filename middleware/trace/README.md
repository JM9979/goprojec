# Trace 包

这个包封装了OpenTelemetry的跟踪功能，提供了简单的API来生成和管理trace ID和span ID。

## 功能

- 创建包含trace ID和span ID的context
- 在现有的trace中创建子span
- 从context中提取trace ID和span ID
- 获取当前span
- 结束span
- HTTP中间件支持，可以在HTTP请求中传递trace信息
- Gin框架中间件支持

## 安装依赖

确保你的项目中安装了必要的依赖：

```bash
go get go.opentelemetry.io/otel
go get go.opentelemetry.io/otel/trace
go get go.opentelemetry.io/otel/sdk
```

## 基本使用方法

### 创建包含trace的context

```go
// 创建一个包含新trace和span的context
ctx := trace.NewContext(context.Background(), "operation-name")
```

### 提取trace ID和span ID

```go
// 从context中提取trace ID和span ID
traceID, spanID := trace.ExtractIDs(ctx)
fmt.Printf("Trace ID: %s\n", traceID)
fmt.Printf("Span ID: %s\n", spanID)
```

### 创建子span

```go
// 在现有的trace中创建一个新的span
childCtx := trace.WithNewSpan(ctx, "child-operation")
```

### 结束span

```go
// 结束context中的span
trace.EndSpan(ctx)

// 或者使用defer
defer trace.EndSpan(ctx)
```

## HTTP支持

### 服务器中间件

在HTTP服务器中使用中间件来自动管理trace：

```go
import (
    "net/http"
    "your-project-path/middleware/trace"
)

func main() {
    // 创建处理程序
    handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // 从请求context中获取trace信息
        traceID, spanID := trace.ExtractIDs(r.Context())
        
        // 业务逻辑...
        fmt.Fprintf(w, "TraceID: %s, SpanID: %s", traceID, spanID)
    })
    
    // 应用trace中间件
    http.Handle("/", trace.HTTPMiddleware(handler))
    http.ListenAndServe(":8080", nil)
}
```

### HTTP客户端中间件

在HTTP客户端上使用中间件，可以在请求中自动添加trace信息：

```go
import (
    "net/http"
    "your-project-path/middleware/trace"
)

func main() {
    // 创建HTTP客户端并应用中间件
    client := &http.Client{}
    client = trace.HTTPClientMiddleware(client)
    
    // 创建包含trace的context
    ctx := trace.NewContext(context.Background(), "api-request")
    
    // 使用context创建请求
    req, _ := http.NewRequestWithContext(ctx, "GET", "https://api.example.com", nil)
    
    // 发送请求（trace信息会自动添加到请求头）
    resp, _ := client.Do(req)
    // ...
}
```

### 手动处理请求头

可以手动将trace信息注入请求头：

```go
// 注入trace信息到HTTP请求
req = trace.InjectTraceToRequest(req)

// 从HTTP请求中提取trace信息
traceID, spanID := trace.ExtractTraceFromRequest(req)
```

## Gin框架支持

### 使用Gin中间件

在Gin框架中使用trace中间件：

```go
import (
    "github.com/gin-gonic/gin"
    "your-project-path/middleware/trace"
)

func main() {
    // 创建Gin引擎
    r := gin.Default()
    
    // 应用trace中间件
    r.Use(trace.GinMiddleware())
    
    // 定义路由
    r.GET("/", func(c *gin.Context) {
        // 从Gin上下文中获取trace信息
        traceID := trace.GetTraceIDFromGin(c)
        spanID := trace.GetSpanIDFromGin(c)
        
        // 业务逻辑...
        c.JSON(200, gin.H{
            "trace_id": traceID,
            "span_id": spanID,
        })
    })
    
    r.Run(":8080")
}
```

### 在Gin处理函数中创建子span

```go
import (
    "github.com/gin-gonic/gin"
    "your-project-path/middleware/trace"
)

func handleRequest(c *gin.Context) {
    // 获取当前trace信息
    traceID := trace.GetTraceIDFromGin(c)
    spanID := trace.GetSpanIDFromGin(c)
    
    // 创建子span进行数据库操作
    c = trace.WithChildSpan(c, "database-operation")
    
    // 执行数据库操作...
    
    // 获取子span的trace信息
    dbTraceID := trace.GetTraceIDFromGin(c)
    dbSpanID := trace.GetSpanIDFromGin(c)
    
    // 响应请求
    c.JSON(200, gin.H{
        "message": "success",
        "trace_id": traceID,
        "span_id": spanID,
        "db_span_id": dbSpanID,
    })
}
```

## 完整示例

详见 `example` 和 `gin_example` 目录中的示例代码。 