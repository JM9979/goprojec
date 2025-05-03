package ft

import (
	"context"
	"fmt"

	"ginproject/entity/ft"
	"ginproject/middleware/log"
)

// GetPoolListByFtContractId 根据代币合约ID获取相关的流动池列表
func (l *FtLogic) GetPoolListByFtContractId(ctx context.Context, req *ft.TBC20PoolListRequest) (*ft.TBC20PoolListResponse, error) {
	// 使用entity层的验证逻辑
	if err := req.Validate(); err != nil {
		log.ErrorWithContextf(ctx, "参数验证失败: %v", err)
		return nil, fmt.Errorf("参数验证失败: %w", err)
	}

	// 查询相关的流动池列表
	log.InfoWithContextf(ctx, "开始查询代币相关流动池: 合约ID=%s", req.FtContractId)
	poolList, err := l.ftPoolNftDAO.GetPoolListByFtContractId(ctx, req.FtContractId)
	if err != nil {
		log.ErrorWithContextf(ctx, "查询流动池列表失败: %v", err)
		return nil, fmt.Errorf("查询流动池列表失败: %w", err)
	}

	// 初始化响应对象
	response := &ft.TBC20PoolListResponse{
		PoolList:       make([]ft.TBC20PoolInfo, 0, len(poolList)),
		TotalPoolCount: int64(len(poolList)),
	}

	// 如果没有找到任何池，返回空列表
	if len(poolList) == 0 {
		log.InfoWithContextf(ctx, "未找到相关流动池: 合约ID=%s", req.FtContractId)
		return response, nil
	}

	// 获取当前代币的信息
	tokenInfo, err := l.ftTokensDAO.GetFtTokenById(req.FtContractId)
	if err != nil {
		log.WarnWithContextf(ctx, "获取代币信息失败: %v", err)
		// 继续处理，使用默认值
	}

	tokenName := "未知"
	if tokenInfo != nil {
		tokenName = tokenInfo.FtName
	}

	// 处理每个流动池
	for _, pool := range poolList {
		poolInfo := ft.TBC20PoolInfo{
			PoolId:              pool.NftContractId,   // 使用NFT合约ID作为池ID
			TokenPairAId:        "TBC",                // TBC作为A代币
			TokenPairAName:      "TBC",                // TBC名称
			TokenPairBId:        req.FtContractId,     // 请求的代币作为B代币
			TokenPairBName:      tokenName,            // 代币名称
			PoolCreateTimestamp: pool.CreateTimestamp, // 创建时间戳
		}

		response.PoolList = append(response.PoolList, poolInfo)
	}

	// 更新总数
	response.TotalPoolCount = int64(len(response.PoolList))

	log.InfoWithContextf(ctx, "查询流动池列表成功: 合约ID=%s, 总数=%d",
		req.FtContractId, response.TotalPoolCount)

	return response, nil
}
