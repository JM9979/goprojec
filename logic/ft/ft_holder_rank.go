package ft

import (
	"context"
	"fmt"

	"ginproject/entity/ft"
	"ginproject/entity/utility"
	"ginproject/middleware/log"
)

// GetFtHolderRank 获取代币持有者排名
func (l *FtLogic) GetFtHolderRank(ctx context.Context, req *ft.FtHolderRankRequest) (*ft.FtHolderRankResponse, error) {
	// 参数合法性校验
	if err := req.Validate(); err != nil {
		log.ErrorWithContextf(ctx, "参数验证失败: %v", err)
		return nil, fmt.Errorf("参数验证失败: %v", err)
	}

	// 解析分页参数
	page, err := req.ParsePage()
	if err != nil {
		log.ErrorWithContextf(ctx, "解析页码失败: %v", err)
		return nil, fmt.Errorf("解析页码失败: %v", err)
	}

	size, err := req.ParseSize()
	if err != nil {
		log.ErrorWithContextf(ctx, "解析每页记录数失败: %v", err)
		return nil, fmt.Errorf("解析每页记录数失败: %v", err)
	}

	log.InfoWithContextf(ctx, "查询代币持有者排名, 合约ID: %s, 页码: %d, 每页记录数: %d",
		req.ContractId, page, size)

	// 查询代币基本信息
	token, err := l.ftTokensDAO.GetFtTokenById(req.ContractId)
	if err != nil {
		log.ErrorWithContextf(ctx, "获取代币信息失败: %v", err)
		return nil, fmt.Errorf("获取代币信息失败: %v", err)
	}

	// 查询代币持有者数量
	holdersCount, err := l.ftBalanceDAO.GetHoldersCountByContractId(req.ContractId)
	if err != nil {
		log.ErrorWithContextf(ctx, "获取代币持有者数量失败: %v", err)
		return nil, fmt.Errorf("获取代币持有者数量失败: %v", err)
	}

	// 查询代币总供应量
	totalSupply := token.FtSupply

	// 查询持有者排名数据
	balances, err := l.ftBalanceDAO.GetFtBalanceRankByContractId(ctx, req.ContractId, page, size)
	if err != nil {
		log.ErrorWithContextf(ctx, "获取代币持有者排名失败: %v", err)
		return nil, fmt.Errorf("获取代币持有者排名失败: %v", err)
	}

	// 构造响应数据
	holderRankList := make([]ft.HolderRankInfo, 0, len(balances))

	// 处理每个持有者的排名信息
	for i, balance := range balances {
		// 计算排名序号（起始页码*每页大小+当前索引+1）
		rank := page*size + i + 1

		// 计算持有比例
		holdRatio := float64(0)
		if totalSupply > 0 {
			holdRatio = float64(balance.FtBalance) / float64(totalSupply)
		}

		// 转换组合脚本为地址
		address := "未知地址"
		if len(balance.FtHolderCombineScript) > 2 && balance.FtHolderCombineScript[len(balance.FtHolderCombineScript)-2:] == "00" {
			// 普通地址
			addr, err := utility.ConvertCombineScriptToAddress(balance.FtHolderCombineScript)
			if err == nil {
				address = addr
			} else {
				log.WarnWithContextf(ctx, "转换地址失败: %s, %v", balance.FtHolderCombineScript, err)
			}
		} else {
			// 其他类型地址（可能是合约或多签）
			address = "Contract_" + balance.FtHolderCombineScript
		}

		// 添加到列表
		holderRankInfo := ft.HolderRankInfo{
			Address:   address,
			Balance:   balance.FtBalance,
			Rank:      rank,
			HoldRatio: holdRatio,
		}
		holderRankList = append(holderRankList, holderRankInfo)
	}

	// 构造返回响应
	response := &ft.FtHolderRankResponse{
		FtContractId:   req.ContractId,
		FtDecimal:      int(token.FtDecimal),
		FtHoldersCount: int(holdersCount),
		HolderRank:     holderRankList,
	}

	log.InfoWithContextf(ctx, "获取代币持有者排名成功, 合约ID: %s, 返回记录数: %d",
		req.ContractId, len(holderRankList))
	return response, nil
}
