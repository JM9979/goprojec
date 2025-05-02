# RPC 测试工具集

本目录包含了用于测试RPC连接的各种工具和测试用例。

## 测试文件说明

### 1. 单元测试

- **client_test.go**：完整的RPC客户端单元测试
  - 测试RPC连接的连通性
  - 测试获取网络信息功能
  - 运行方式：`go test -v ./repo/rpc/test`

- **connectivity_test.go**：简化的连通性测试
  - 专注于连通性测试的简化版本
  - 运行方式：`go test -v -run TestConnectivity ./repo/rpc/test`

### 2. 命令行工具

- **check_connectivity.go**：快速检查RPC连通性的命令行工具
  - 简单的连通性检查，带有清晰的步骤显示
  - 运行方式：`go run repo/rpc/test/check_connectivity.go`

- **rpc_test_cmd.go**：全面的RPC测试命令行工具
  - 测试同步和异步RPC调用
  - 获取更多节点信息
  - 运行方式：`go run repo/rpc/test/rpc_test_cmd.go`

## 使用场景

- **开发测试**：使用单元测试进行RPC客户端开发和调试
- **运维检查**：使用命令行工具快速检查连通性
- **自动化测试**：通过CI/CD系统调用测试用例进行自动化检查

## 运行示例

1. 运行单元测试：
```
go test -v ./repo/rpc/test
```

2. 运行连通性检查工具：
```
go run repo/rpc/test/check_connectivity.go
```

## 注意事项

- 测试前需确保配置文件中RPC相关配置正确设置
- 测试需要目标RPC节点可访问
- 测试报告会同时输出到控制台和日志文件 