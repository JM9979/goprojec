package electrumx

import (
	"context"
	"testing"
	"time"

	"ginproject/middleware/log"
)

// setupTestClient 设置测试客户端
func setupTestClient(t *testing.T) *ElectrumXClient {
	// 配置测试环境
	testConfig := &RPCConfig{
		Host:     "testnet.electrumx-server.com", // 使用测试网络服务器
		Port:     50001,                          // 标准端口
		UseTLS:   false,
		Timeout:  10 * time.Second,
		Protocol: "TCP",
	}

	// 更新配置
	if err := UpdateConfig(testConfig); err != nil {
		t.Fatalf("更新配置失败: %v", err)
	}

	// 创建客户端
	client, err := NewClient()
	if err != nil {
		t.Fatalf("创建客户端失败: %v", err)
	}

	return client
}

// TestConnection 测试连接
func TestConnection(t *testing.T) {
	// 跳过测试，避免真实网络调用
	t.Skip("跳过网络测试")

	client := setupTestClient(t)

	// 测试连接
	err := client.Connect()
	if err != nil {
		t.Fatalf("连接失败: %v", err)
	}
	defer client.Disconnect()

	t.Log("连接成功")
}

// TestServerVersion 测试获取服务器版本
func TestServerVersion(t *testing.T) {
	// 跳过测试，避免真实网络调用
	t.Skip("跳过网络测试")

	client := setupTestClient(t)

	// 调用RPC方法
	result, err := client.CallRPC("server.version", []interface{}{"test-client", "1.4"})
	if err != nil {
		t.Fatalf("调用RPC失败: %v", err)
	}

	t.Logf("服务器版本: %s", string(result))
}

// TestAsyncCall 测试异步调用
func TestAsyncCall(t *testing.T) {
	// 跳过测试，避免真实网络调用
	t.Skip("跳过网络测试")

	client := setupTestClient(t)

	// 创建上下文
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 调用异步RPC方法
	resultChan := client.CallRPCAsync(ctx, "server.ping", []interface{}{})

	// 等待结果
	select {
	case result := <-resultChan:
		if result.Error != nil {
			t.Fatalf("异步调用失败: %v", result.Error)
		}
		t.Logf("异步调用成功: %s", string(result.Result))
	case <-time.After(10 * time.Second):
		t.Fatal("异步调用超时")
	}
}

// TestBlockHeader 测试获取区块头
func TestBlockHeader(t *testing.T) {
	// 跳过测试，避免真实网络调用
	t.Skip("跳过网络测试")

	// 使用全局函数
	if err := Init(); err != nil {
		t.Fatalf("初始化失败: %v", err)
	}

	// 获取区块头
	header, err := GetBlockHeader(100000)
	if err != nil {
		t.Fatalf("获取区块头失败: %v", err)
	}

	t.Logf("区块头: %v", header)
}

// PrintElectrumXStatus 打印ElectrumX状态
func PrintElectrumXStatus(t *testing.T) {
	// 跳过测试，避免真实网络调用
	t.Skip("跳过网络测试")

	// 初始化
	if err := Init(); err != nil {
		t.Fatalf("初始化失败: %v", err)
	}

	// 获取服务器版本
	version, err := ServerVersion()
	if err != nil {
		t.Logf("获取服务器版本失败: %v", err)
	} else {
		t.Logf("服务器版本: %v", version)
	}

	// 获取服务器特性
	features, err := ServerFeatures()
	if err != nil {
		t.Logf("获取服务器特性失败: %v", err)
	} else {
		t.Logf("服务器特性: %v", features)
	}

	// 获取对等节点
	peers, err := ServerPeers()
	if err != nil {
		t.Logf("获取对等节点失败: %v", err)
	} else {
		t.Logf("对等节点数量: %d", len(peers))
	}
}

// 辅助函数 - 打印JSON
func PrintJSON(t *testing.T, label string, data interface{}) {
	log.Debugf("%s: %v", label, data)
	t.Logf("%s: %v", label, data)
}
