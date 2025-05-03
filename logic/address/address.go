package address

import (
	"context"
	"fmt"

	"ginproject/entity/electrumx"
	"ginproject/middleware/log"
	rpcex "ginproject/repo/rpc/electrumx"
	utility "ginproject/entity/utility"
	"go.uber.org/zap"
)

// AddressLogic 地址业务逻辑
type AddressLogic struct{}

// NewAddressLogic 创建地址业务逻辑实例
func NewAddressLogic() *AddressLogic {
	return &AddressLogic{}
}

// GetAddressUnspentUtxos 获取地址的未花费交易输出
func (l *AddressLogic) GetAddressUnspentUtxos(ctx context.Context, address string) (electrumx.UtxoResponse, error) {
	// 记录开始处理的日志
	log.InfoWithContext(ctx, "开始获取地址的UTXO", zap.String("address", address))

	// 验证地址合法性
	valid, addrType, err := utility.ValidateWIFAddress(address)
	if err != nil || !valid {
		log.ErrorWithContext(ctx, "地址验证失败", zap.String("address", address), zap.Error(err))
		return nil, fmt.Errorf("无效的地址格式: %w", err)
	}

	log.InfoWithContext(ctx, "地址验证通过", zap.String("address", address), zap.Int("type", addrType))

	// 将地址转换为脚本哈希
	scriptHash, err := rpcex.AddressToScriptHash(address)
	if err != nil {
		log.ErrorWithContext(ctx, "地址转换为脚本哈希失败", zap.String("address", address), zap.Error(err))
		return nil, fmt.Errorf("地址转换失败: %w", err)
	}

	log.InfoWithContext(ctx, "地址已转换为脚本哈希", zap.String("address", address), zap.String("scriptHash", scriptHash))

	// 调用RPC获取UTXO列表
	utxos, err := rpcex.GetListUnspent(scriptHash)
	if err != nil {
		log.ErrorWithContext(ctx, "获取UTXO失败",
			zap.String("address", address),
			zap.String("scriptHash", scriptHash),
			zap.Error(err))
		return nil, fmt.Errorf("获取UTXO失败: %w", err)
	}

	log.InfoWithContext(ctx, "成功获取地址UTXO", zap.String("address", address), zap.Int("count", len(utxos)))
	return utxos, nil
}
