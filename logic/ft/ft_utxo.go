package ft

import (
	"context"
	"fmt"

	"ginproject/entity/ft"
	"ginproject/middleware/log"
)

// GetFtUtxosByAddress 获取FT UTXO列表
func (l *FtLogic) GetFtUtxosByAddress(ctx context.Context, req *ft.FtUtxoAddressRequest) (*ft.FtUtxoAddressResponse, error) {
	// 使用entity层的验证逻辑
	if err := req.Validate(); err != nil {
		log.ErrorWithContextf(ctx, "参数验证失败: %v", err)
		return nil, fmt.Errorf("参数验证失败: %v", err)
	}

	// 从请求对象中获取组合脚本
	combineScript, err := req.GetCombineScript()
	if err != nil {
		log.ErrorWithContextf(ctx, "获取组合脚本失败: %v", err)
		return nil, fmt.Errorf("获取组合脚本失败: %v", err)
	}

	// 添加00作为校验
	combineScript += "00"

	// 获取代币小数位数
	ftDecimal, err := l.ftTokensDAO.GetFtDecimalByContractId(ctx, req.ContractId)
	if err != nil {
		log.ErrorWithContextf(ctx, "获取代币小数位数失败: %v", err)
		return nil, fmt.Errorf("获取代币小数位数失败: %v", err)
	}

	// 调用DAO层获取未花费的UTXO列表
	utxos, err := l.ftTxoDAO.GetUnspentFtTxosByHolderAndContract(combineScript, req.ContractId)
	if err != nil {
		log.ErrorWithContextf(ctx, "查询FT UTXO列表失败: %v", err)
		return nil, fmt.Errorf("查询FT UTXO列表失败: %v", err)
	}

	// 构造响应
	response := &ft.FtUtxoAddressResponse{
		FtUtxoList: make([]*ft.FtUtxoItem, 0, len(utxos)),
	}

	// 遍历UTXO列表，构造UTXO项
	for _, utxo := range utxos {
		utxoItem := &ft.FtUtxoItem{
			UtxoId:       utxo.UtxoTxid,
			UtxoVout:     utxo.UtxoVout,
			UtxoBalance:  utxo.UtxoBalance,
			FtContractId: utxo.FtContractId,
			FtDecimal:    int(ftDecimal),
			FtBalance:    utxo.FtBalance,
		}
		response.FtUtxoList = append(response.FtUtxoList, utxoItem)
	}

	log.InfoWithContextf(ctx, "FT UTXO查询成功: 共%d条记录", len(response.FtUtxoList))

	return response, nil
}
