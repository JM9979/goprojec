package ft

import (
	"context"
	"fmt"

	"ginproject/entity/ft"
	"ginproject/middleware/log"
	"ginproject/repo/db/ft_tokens_dao"
	"ginproject/repo/db/ft_txo_dao"
)

// FtLogic FT代币逻辑处理
type FtLogic struct {
	ftTxoDAO    *ft_txo_dao.FtTxoDAO
	ftTokensDAO *ft_tokens_dao.FtTokensDAO
}

// NewFtLogic 创建FtLogic实例
func NewFtLogic() *FtLogic {
	return &FtLogic{
		ftTxoDAO:    ft_txo_dao.NewFtTxoDAO(),
		ftTokensDAO: ft_tokens_dao.NewFtTokensDAO(),
	}
}

// GetFtBalance 获取FT余额
func (l *FtLogic) GetFtBalance(ctx context.Context, req *ft.FtBalanceAddressRequest) (*ft.FtBalanceAddressResponse, error) {
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

	// 调用DAO层获取余额
	ftBalance, err := l.ftTxoDAO.GetTotalBalanceByHolder(ctx, combineScript, req.ContractId)
	if err != nil {
		log.ErrorWithContextf(ctx, "查询FT余额失败: %v", err)
		return nil, fmt.Errorf("查询FT余额失败: %v", err)
	}

	log.InfoWithContextf(ctx, "FT余额查询成功: %d", ftBalance)

	// 构造响应
	response := &ft.FtBalanceAddressResponse{
		CombineScript: combineScript,
		FtContractId:  req.ContractId,
		FtDecimal:     int(ftDecimal),
		FtBalance:     ftBalance,
	}

	return response, nil
}
