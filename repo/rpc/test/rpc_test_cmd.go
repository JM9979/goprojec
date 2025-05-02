//go:build ignore
// +build ignore

package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"ginproject/middleware/conf"
	"ginproject/middleware/log"
	"ginproject/repo/rpc"
)

// 这是一个简单的命令行工具，用于测试RPC连接
// 可以通过 go run repo/rpc/test/rpc_test_cmd.go 运行
func main() {
	// 初始化配置
	fmt.Println("加载配置...")
	err := conf.LoadConfig()
	if err != nil {
		fmt.Printf("加载配置失败: %v\n", err)
		os.Exit(1)
	}

	// 初始化日志
	fmt.Println("初始化日志...")
	logConfig := conf.GetLogConfig()
	err = log.InitLogger(logConfig, "rpc-test-cmd")
	if err != nil {
		fmt.Printf("初始化日志失败: %v\n", err)
		os.Exit(1)
	}

	// 初始化RPC客户端
	fmt.Println("初始化RPC客户端...")
	err = rpc.Init()
	if err != nil {
		fmt.Printf("初始化RPC客户端失败: %v\n", err)
		os.Exit(1)
	}

	// 显示RPC配置
	fmt.Printf("RPC配置: URL=%s, 用户=%s\n", rpc.Config.URL, rpc.Config.User)

	// 测试同步调用
	fmt.Println("\n== 测试同步RPC调用 ==")
	testSyncRPC()

	// 测试异步调用
	fmt.Println("\n== 测试异步RPC调用 ==")
	testAsyncRPC()

	fmt.Println("\n所有测试完成！")
}

// 测试同步RPC调用
func testSyncRPC() {
	fmt.Println("调用 getblockcount...")
	result, err := rpc.CallNodeRPC("getblockcount", []interface{}{}, false)
	if err != nil {
		fmt.Printf("RPC调用失败: %v\n", err)
		return
	}

	blockHeight, ok := result.(float64)
	if !ok {
		fmt.Printf("无法解析区块高度，收到的数据类型: %T\n", result)
		return
	}

	fmt.Printf("当前区块高度: %.0f\n", blockHeight)

	// 获取网络信息
	fmt.Println("\n调用 getnetworkinfo...")
	result, err = rpc.CallNodeRPC("getnetworkinfo", []interface{}{}, false)
	if err != nil {
		fmt.Printf("获取网络信息失败: %v\n", err)
		return
	}

	networkInfo, ok := result.(map[string]interface{})
	if !ok {
		fmt.Printf("无法解析网络信息，收到的数据类型: %T\n", result)
		return
	}

	// 输出一些关键信息
	fmt.Printf("节点版本: %v\n", networkInfo["version"])
	fmt.Printf("客户端标识: %v\n", networkInfo["subversion"])
	fmt.Printf("连接数: %v\n", networkInfo["connections"])
}

// 测试异步RPC调用
func testAsyncRPC() {
	// 创建带有超时的上下文
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	fmt.Println("异步调用 getblockcount...")
	resultChan := rpc.CallNodeRPCAsync(ctx, "getblockcount", []interface{}{}, false)

	// 等待结果
	result := <-resultChan
	if result.Error != nil {
		fmt.Printf("异步RPC调用失败: %v\n", result.Error)
		return
	}

	blockHeight, ok := result.Result.(float64)
	if !ok {
		fmt.Printf("无法解析区块高度，收到的数据类型: %T\n", result.Result)
		return
	}

	fmt.Printf("当前区块高度 (异步): %.0f\n", blockHeight)
}
