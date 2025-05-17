package ft

import (
	"context"
	"fmt"
	"sync"

	"ginproject/entity/dbtable"
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

	// 使用协程并发执行三个数据库查询操作
	var wg sync.WaitGroup
	wg.Add(3)

	// 用于存储查询结果和错误
	var token *dbtable.FtTokens
	var tokenErr error

	var holdersCount int64
	var holdersCountErr error

	var balances []*dbtable.FtBalance
	var balancesErr error

	// 查询代币基本信息
	go func() {
		defer wg.Done()
		token, tokenErr = l.ftTokensDAO.GetFtTokenById(req.ContractId)
		if tokenErr != nil {
			log.ErrorWithContextf(ctx, "获取代币信息失败: %v", tokenErr)
		}
	}()

	// 查询代币持有者数量
	go func() {
		defer wg.Done()
		holdersCount, holdersCountErr = l.ftBalanceDAO.GetHoldersCountByContractId(req.ContractId)
		if holdersCountErr != nil {
			log.ErrorWithContextf(ctx, "获取代币持有者数量失败: %v", holdersCountErr)
		}
	}()

	// 查询持有者排名数据
	go func() {
		defer wg.Done()
		balances, balancesErr = l.ftBalanceDAO.GetFtBalanceRankByContractId(ctx, req.ContractId, page, size)
		if balancesErr != nil {
			log.ErrorWithContextf(ctx, "获取代币持有者排名失败: %v", balancesErr)
		}
	}()

	// 等待所有协程完成
	wg.Wait()

	// 检查错误
	if tokenErr != nil {
		return nil, fmt.Errorf("获取代币信息失败: %v", tokenErr)
	}

	if holdersCountErr != nil {
		return nil, fmt.Errorf("获取代币持有者数量失败: %v", holdersCountErr)
	}

	if balancesErr != nil {
		return nil, fmt.Errorf("获取代币持有者排名失败: %v", balancesErr)
	}

	// 查询代币总供应量
	totalSupply := token.FtSupply

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
