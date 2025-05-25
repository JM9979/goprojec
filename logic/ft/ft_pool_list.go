package ft

import (
	"context"
	"fmt"
	"sync"
	"time"

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

// GetAllPoolList 获取所有交易池列表
func (l *FtLogic) GetAllPoolList(ctx context.Context, req *ft.TBC20PoolPageRequest) (*ft.TBC20PoolPageResponse, error) {
	// 使用entity层的验证逻辑
	if err := req.Validate(); err != nil {
		log.ErrorWithContextf(ctx, "参数验证失败: %v", err)
		return nil, fmt.Errorf("参数验证失败: %w", err)
	}

	// 记录请求开始日志
	log.InfoWithContextf(ctx, "开始获取所有流动池列表: 页码=%d, 页大小=%d", req.Page, req.Size)

	// 使用异步查询方法获取流动池列表
	resultChan, err := l.ftPoolNftDAO.GetAllPoolsWithPagination(ctx, req.Page, req.Size)
	if err != nil {
		log.ErrorWithContextf(ctx, "启动异步查询流动池列表失败: %v", err)
		return nil, fmt.Errorf("启动异步查询流动池列表失败: %w", err)
	}

	// 等待异步查询结果
	log.InfoWithContextf(ctx, "等待异步查询结果...")
	result := <-resultChan

	// 检查查询是否成功
	if result.Error != nil {
		log.ErrorWithContextf(ctx, "异步查询流动池列表失败: %v", result.Error)
		return nil, fmt.Errorf("异步查询流动池列表失败: %w", result.Error)
	}

	// 初始化响应对象
	response := &ft.TBC20PoolPageResponse{
		PoolList:       make([]ft.TBC20PoolInfo, len(result.Results)),
		TotalPoolCount: result.TotalCount,
	}

	// 如果没有找到任何池，返回空列表
	if len(result.Results) == 0 {
		log.InfoWithContextf(ctx, "未找到流动池信息，返回空列表")
		return response, nil
	}

	// 收集所有需要查询的代币IDs
	tokenIds := make([]string, 0, len(result.Results))
	tokenIdMap := make(map[string]bool)

	for _, pool := range result.Results {
		if pool.TokenContractId != "" && !tokenIdMap[pool.TokenContractId] {
			tokenIds = append(tokenIds, pool.TokenContractId)
			tokenIdMap[pool.TokenContractId] = true
		}
	}

	// 使用并发方式批量查询代币信息
	type tokenInfoResult struct {
		TokenId   string
		TokenName string
		Error     error
	}

	var wg sync.WaitGroup
	tokenInfoChan := make(chan tokenInfoResult, len(tokenIds))
	tokenNameCache := make(map[string]string)

	// 设置代币查询超时上下文
	tokenCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	// 启动并发查询代币信息
	for _, tokenId := range tokenIds {
		wg.Add(1)
		go func(id string) {
			defer wg.Done()

			select {
			case <-tokenCtx.Done():
				// 上下文已取消
				tokenInfoChan <- tokenInfoResult{
					TokenId:   id,
					TokenName: "未知",
					Error:     tokenCtx.Err(),
				}
				log.WarnWithContextf(ctx, "代币信息查询超时: 代币ID=%s", id)
			default:
				// 查询代币信息
				tokenInfo, err := l.ftTokensDAO.GetFtTokenById(id)
				if err != nil {
					tokenInfoChan <- tokenInfoResult{
						TokenId:   id,
						TokenName: "未知",
						Error:     err,
					}
					log.WarnWithContextf(ctx, "获取代币信息失败: 代币ID=%s, 错误=%v", id, err)
				} else {
					name := "未知"
					if tokenInfo != nil {
						name = tokenInfo.FtName
					}
					tokenInfoChan <- tokenInfoResult{
						TokenId:   id,
						TokenName: name,
						Error:     nil,
					}
				}
			}
		}(tokenId)
	}

	// 使用另一个goroutine等待所有查询完成并关闭通道
	go func() {
		wg.Wait()
		close(tokenInfoChan)
	}()

	// 收集代币查询结果
	for result := range tokenInfoChan {
		if result.Error == nil {
			tokenNameCache[result.TokenId] = result.TokenName
		} else {
			log.WarnWithContextf(ctx, "代币[%s]信息查询失败: %v", result.TokenId, result.Error)
			tokenNameCache[result.TokenId] = "未知"
		}
	}

	// 组装池信息
	for i, pool := range result.Results {
		tokenName := "未知"
		if name, ok := tokenNameCache[pool.TokenContractId]; ok {
			tokenName = name
		}

		response.PoolList[i] = ft.TBC20PoolInfo{
			PoolId:              pool.NftContractId,
			TokenPairAId:        "TBC",
			TokenPairAName:      "TBC",
			TokenPairBId:        pool.TokenContractId,
			TokenPairBName:      tokenName,
			PoolCreateTimestamp: pool.CreateTimestamp,
		}
	}

	log.InfoWithContextf(ctx, "获取所有流动池列表成功: 总数=%d, 本页获取数=%d",
		response.TotalPoolCount, len(response.PoolList))

	return response, nil
}
