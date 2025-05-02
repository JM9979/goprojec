package test

import (
	"fmt"
	"os"
	"testing"

	"ginproject/middleware/conf"
	"ginproject/middleware/log"
	"ginproject/repo/rpc"

	"go.uber.org/zap"
)

func TestMain(m *testing.M) {
	// 初始化配置
	err := conf.LoadConfig()
	if err != nil {
		fmt.Printf("加载配置失败: %v\n", err)
		os.Exit(1)
	}

	// 初始化日志
	logConfig := conf.GetLogConfig()
	err = log.InitLogger(logConfig, "rpc-test")
	if err != nil {
		fmt.Printf("初始化日志失败: %v\n", err)
		os.Exit(1)
	}

	// 初始化RPC客户端
	err = rpc.Init()
	if err != nil {
		fmt.Printf("初始化RPC客户端失败: %v\n", err)
		os.Exit(1)
	}

	// 运行测试
	code := m.Run()
	os.Exit(code)
}

// TestRPCConnectivity 测试RPC连接的连通性
func TestRPCConnectivity(t *testing.T) {
	// 尝试调用一个简单的RPC方法，例如getblockcount
	t.Log("开始测试RPC连通性...")
	log.Info("开始测试RPC连通性")

	// 记录当前配置信息
	t.Logf("当前RPC配置 - URL: %s, 用户: %s",
		rpc.Config.URL, rpc.Config.User)
	log.Info("当前RPC配置信息",
		zap.String("url", rpc.Config.URL),
		zap.String("user", rpc.Config.User))

	// 调用RPC方法
	result, err := rpc.CallNodeRPC("getblockcount", []interface{}{}, false)
	if err != nil {
		t.Errorf("RPC调用失败: %v", err)
		log.Error("RPC调用失败", zap.Error(err))
		return
	}

	// 检查结果
	blockHeight, ok := result.(float64)
	if !ok {
		t.Errorf("无法解析区块高度，收到的数据类型不是数字: %T", result)
		log.Error("无法解析区块高度", zap.Any("result", result))
		return
	}

	t.Logf("连接成功! 当前区块高度: %.0f", blockHeight)
	log.Info("RPC连接成功", zap.Float64("blockHeight", blockHeight))
}

// TestRPCGetNetworkInfo 测试获取网络信息
func TestRPCGetNetworkInfo(t *testing.T) {
	t.Log("开始测试获取网络信息...")
	log.Info("开始测试获取网络信息")

	// 调用RPC方法
	result, err := rpc.CallNodeRPC("getnetworkinfo", []interface{}{}, false)
	if err != nil {
		t.Errorf("获取网络信息失败: %v", err)
		log.Error("获取网络信息失败", zap.Error(err))
		return
	}

	// 检查结果是否为映射
	networkInfo, ok := result.(map[string]interface{})
	if !ok {
		t.Errorf("无法解析网络信息，收到的数据类型不是映射: %T", result)
		log.Error("无法解析网络信息", zap.Any("result", result))
		return
	}

	// 输出一些关键信息
	version, versionExists := networkInfo["version"]
	subversion, subversionExists := networkInfo["subversion"]
	connections, connectionsExists := networkInfo["connections"]

	if versionExists && subversionExists && connectionsExists {
		t.Logf("节点版本: %v, 客户端标识: %v, 连接数: %v",
			version, subversion, connections)
		log.Info("获取网络信息成功",
			zap.Any("version", version),
			zap.Any("subversion", subversion),
			zap.Any("connections", connections))
	} else {
		t.Log("节点信息获取成功，但部分字段缺失")
		log.Info("节点信息获取成功，但部分字段缺失",
			zap.Any("networkInfo", networkInfo))
	}
}
