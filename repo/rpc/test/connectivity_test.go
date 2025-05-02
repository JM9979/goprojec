package test

import (
	"testing"

	"ginproject/repo/rpc"
)

// TestConnectivity 测试RPC连通性，不依赖TestMain
func TestConnectivity(t *testing.T) {
	t.Log("开始测试RPC连通性，直接调用URL")

	// 创建一个临时的RPC客户端，不依赖全局配置
	tempConfig := rpc.Config
	t.Logf("使用配置: URL=%s, User=%s", tempConfig.URL, tempConfig.User)

	// 使用一个简单的HTTP GET请求检查URL是否可达
	t.Log("测试URL可达性...")
	result, err := rpc.CallNodeRPC("getblockcount", []interface{}{}, false)
	if err != nil {
		t.Errorf("连通性测试失败: %v", err)
		return
	}

	// 打印结果
	t.Logf("连通性测试成功，响应: %v", result)
}

// 如果需要在命令行直接运行，可以使用以下方式:
// go test -v -run TestConnectivity ./repo/rpc/test
