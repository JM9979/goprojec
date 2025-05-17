package ft

import (
	"context"
	"fmt"
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
		PoolList:       make([]ft.TBC20PoolInfo, 0, len(result.Results)),
		TotalPoolCount: result.TotalCount,
	}

	// 如果没有找到任何池，返回空列表
	if len(result.Results) == 0 {
		log.InfoWithContextf(ctx, "未找到流动池信息，返回空列表")
		return response, nil
	}

	// 处理每个流动池信息
	for _, pool := range result.Results {
		// 初始化代币B的名称
		var tokenPairBName string

		// 只有当TokenContractId不为空时才查询代币信息
		if pool.TokenContractId != "" {
			// 获取代币B的信息
			tokenInfo, err := l.ftTokensDAO.GetFtTokenById(pool.TokenContractId)
			if err != nil {
				log.WarnWithContextf(ctx, "获取代币信息失败: 代币ID=%s, 错误=%v",
					pool.TokenContractId, err)
				// 出错时设置为空字符串，与Python代码行为一致
				tokenPairBName = ""
			} else if tokenInfo != nil {
				// 如果获取到代币信息，使用代币名称
				tokenPairBName = tokenInfo.FtName
			}
		}

		// 构建池信息
		poolInfo := ft.TBC20PoolInfo{
			PoolId:              pool.NftContractId,
			TokenPairAId:        "TBC",
			TokenPairAName:      "TBC",
			TokenPairBId:        pool.TokenContractId,
			TokenPairBName:      tokenPairBName,
			PoolCreateTimestamp: pool.CreateTimestamp,
		}

		// 添加到响应列表
		response.PoolList = append(response.PoolList, poolInfo)
	}

	log.InfoWithContextf(ctx, "处理流动池列表完成: 总数=%d, 返回数量=%d",
		response.TotalPoolCount, len(response.PoolList))

	return response, nil
}

// GetAllPoolListWithTimeout 获取所有交易池列表（带超时处理）
func (l *FtLogic) GetAllPoolListWithTimeout(ctx context.Context, req *ft.TBC20PoolPageRequest, timeout time.Duration) (*ft.TBC20PoolPageResponse, error) {
	// 使用entity层的验证逻辑
	if err := req.Validate(); err != nil {
		log.ErrorWithContextf(ctx, "参数验证失败: %v", err)
		return nil, fmt.Errorf("参数验证失败: %w", err)
	}

	// 创建带超时的上下文
	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// 记录请求开始日志
	log.InfoWithContextf(ctx, "开始获取所有流动池列表（超时=%s）: 页码=%d, 页大小=%d",
		timeout.String(), req.Page, req.Size)

	// 使用异步查询方法获取流动池列表
	resultChan, err := l.ftPoolNftDAO.GetAllPoolsWithPagination(timeoutCtx, req.Page, req.Size)
	if err != nil {
		log.ErrorWithContextf(ctx, "启动异步查询流动池列表失败: %v", err)
		return nil, fmt.Errorf("启动异步查询流动池列表失败: %w", err)
	}

	// 等待异步查询结果，使用select处理可能的超时情况
	log.InfoWithContextf(ctx, "等待异步查询结果，最多等待%s...", timeout.String())

	var result struct {
		Results []struct {
			NftContractId   string `gorm:"column:nft_contract_id"`
			CreateTimestamp int64  `gorm:"column:nft_create_timestamp"`
			TokenContractId string `gorm:"column:nft_icon"`
		}
		TotalCount int64
		Error      error
	}

	select {
	case res, ok := <-resultChan:
		if !ok {
			log.ErrorWithContextf(ctx, "异步查询通道已关闭")
			return nil, fmt.Errorf("异步查询通道已关闭")
		}
		result = res
	case <-timeoutCtx.Done():
		log.ErrorWithContextf(ctx, "异步查询流动池列表超时: %v", timeoutCtx.Err())
		return nil, fmt.Errorf("查询流动池列表超时: %w", timeoutCtx.Err())
	}

	// 检查查询是否成功
	if result.Error != nil {
		log.ErrorWithContextf(ctx, "异步查询流动池列表失败: %v", result.Error)
		return nil, fmt.Errorf("异步查询流动池列表失败: %w", result.Error)
	}

	// 初始化响应对象
	response := &ft.TBC20PoolPageResponse{
		PoolList:       make([]ft.TBC20PoolInfo, 0, len(result.Results)),
		TotalPoolCount: result.TotalCount,
	}

	// 如果没有找到任何池，返回空列表
	if len(result.Results) == 0 {
		log.InfoWithContextf(ctx, "未找到流动池信息，返回空列表")
		return response, nil
	}

	// 处理每个流动池信息
	for _, pool := range result.Results {
		// 初始化代币B的名称
		var tokenPairBName string

		// 只有当TokenContractId不为空时才查询代币信息
		if pool.TokenContractId != "" {
			// 获取代币B的信息
			tokenInfo, err := l.ftTokensDAO.GetFtTokenById(pool.TokenContractId)
			if err != nil {
				log.WarnWithContextf(ctx, "获取代币信息失败: 代币ID=%s, 错误=%v",
					pool.TokenContractId, err)
				// 出错时设置为空字符串，与Python代码行为一致
				tokenPairBName = ""
			} else if tokenInfo != nil {
				// 如果获取到代币信息，使用代币名称
				tokenPairBName = tokenInfo.FtName
			}
		}

		// 构建池信息
		poolInfo := ft.TBC20PoolInfo{
			PoolId:              pool.NftContractId,
			TokenPairAId:        "TBC",
			TokenPairAName:      "TBC",
			TokenPairBId:        pool.TokenContractId,
			TokenPairBName:      tokenPairBName,
			PoolCreateTimestamp: pool.CreateTimestamp,
		}

		// 添加到响应列表
		response.PoolList = append(response.PoolList, poolInfo)
	}

	log.InfoWithContextf(ctx, "处理流动池列表完成: 总数=%d, 返回数量=%d",
		response.TotalPoolCount, len(response.PoolList))

	return response, nil
}

// GetAllPoolListParallel 使用WaitGroup并行优化的流动池查询
func (l *FtLogic) GetAllPoolListParallel(ctx context.Context, req *ft.TBC20PoolPageRequest) (*ft.TBC20PoolPageResponse, error) {
	// 使用entity层的验证逻辑
	if err := req.Validate(); err != nil {
		log.ErrorWithContextf(ctx, "参数验证失败: %v", err)
		return nil, fmt.Errorf("参数验证失败: %w", err)
	}

	// 记录请求开始日志
	log.InfoWithContextf(ctx, "开始WaitGroup并行查询流动池列表: 页码=%d, 页大小=%d", req.Page, req.Size)

	// 使用WaitGroup并行优化的查询方法
	resultChan, err := l.ftPoolNftDAO.GetAllPoolsWithPaginationParallel(ctx, req.Page, req.Size)
	if err != nil {
		log.ErrorWithContextf(ctx, "启动WaitGroup并行查询失败: %v", err)
		return nil, fmt.Errorf("启动WaitGroup并行查询失败: %w", err)
	}

	// 等待异步查询结果
	log.InfoWithContextf(ctx, "等待WaitGroup并行查询结果...")
	result := <-resultChan

	// 检查查询是否成功
	if result.Error != nil {
		log.ErrorWithContextf(ctx, "WaitGroup并行查询失败: %v", result.Error)
		return nil, fmt.Errorf("WaitGroup并行查询失败: %w", result.Error)
	}

	// 初始化响应对象
	response := &ft.TBC20PoolPageResponse{
		PoolList:       make([]ft.TBC20PoolInfo, 0, len(result.Results)),
		TotalPoolCount: result.TotalCount,
	}

	// 如果没有找到任何池，返回空列表
	if len(result.Results) == 0 {
		log.InfoWithContextf(ctx, "未找到流动池信息，返回空列表")
		return response, nil
	}

	// 处理每个流动池信息
	for _, pool := range result.Results {
		// 初始化代币B的名称
		var tokenPairBName string

		// 只有当TokenContractId不为空时才查询代币信息
		if pool.TokenContractId != "" {
			// 获取代币B的信息
			tokenInfo, err := l.ftTokensDAO.GetFtTokenById(pool.TokenContractId)
			if err != nil {
				log.WarnWithContextf(ctx, "获取代币信息失败: 代币ID=%s, 错误=%v",
					pool.TokenContractId, err)
				// 出错时设置为空字符串，与Python代码行为一致
				tokenPairBName = ""
			} else if tokenInfo != nil {
				// 如果获取到代币信息，使用代币名称
				tokenPairBName = tokenInfo.FtName
			}
		}

		// 构建池信息
		poolInfo := ft.TBC20PoolInfo{
			PoolId:              pool.NftContractId,
			TokenPairAId:        "TBC",
			TokenPairAName:      "TBC",
			TokenPairBId:        pool.TokenContractId,
			TokenPairBName:      tokenPairBName,
			PoolCreateTimestamp: pool.CreateTimestamp,
		}

		// 添加到响应列表
		response.PoolList = append(response.PoolList, poolInfo)
	}

	log.InfoWithContextf(ctx, "WaitGroup并行查询完成: 总数=%d, 返回数量=%d",
		response.TotalPoolCount, len(response.PoolList))

	return response, nil
}

// GetAllPoolListParallelWithTimeout 使用WaitGroup并行优化，带超时的流动池查询
func (l *FtLogic) GetAllPoolListParallelWithTimeout(ctx context.Context, req *ft.TBC20PoolPageRequest, timeout time.Duration) (*ft.TBC20PoolPageResponse, error) {
	// 使用entity层的验证逻辑
	if err := req.Validate(); err != nil {
		log.ErrorWithContextf(ctx, "参数验证失败: %v", err)
		return nil, fmt.Errorf("参数验证失败: %w", err)
	}

	// 创建带超时的上下文
	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// 记录请求开始日志
	log.InfoWithContextf(ctx, "开始WaitGroup并行查询流动池列表（超时=%s）: 页码=%d, 页大小=%d",
		timeout.String(), req.Page, req.Size)

	// 使用WaitGroup并行优化的查询方法
	resultChan, err := l.ftPoolNftDAO.GetAllPoolsWithPaginationParallel(timeoutCtx, req.Page, req.Size)
	if err != nil {
		log.ErrorWithContextf(ctx, "启动WaitGroup并行查询失败: %v", err)
		return nil, fmt.Errorf("启动WaitGroup并行查询失败: %w", err)
	}

	// 等待异步查询结果，使用select处理可能的超时情况
	log.InfoWithContextf(ctx, "等待WaitGroup并行查询结果，最多等待%s...", timeout.String())

	var result struct {
		Results []struct {
			NftContractId   string `gorm:"column:nft_contract_id"`
			CreateTimestamp int64  `gorm:"column:nft_create_timestamp"`
			TokenContractId string `gorm:"column:nft_icon"`
		}
		TotalCount int64
		Error      error
	}

	select {
	case res, ok := <-resultChan:
		if !ok {
			log.ErrorWithContextf(ctx, "WaitGroup并行查询通道已关闭")
			return nil, fmt.Errorf("WaitGroup并行查询通道已关闭")
		}
		result = res
	case <-timeoutCtx.Done():
		log.ErrorWithContextf(ctx, "WaitGroup并行查询超时: %v", timeoutCtx.Err())
		return nil, fmt.Errorf("WaitGroup并行查询超时: %w", timeoutCtx.Err())
	}

	// 检查查询是否成功
	if result.Error != nil {
		log.ErrorWithContextf(ctx, "WaitGroup并行查询失败: %v", result.Error)
		return nil, fmt.Errorf("WaitGroup并行查询失败: %w", result.Error)
	}

	// 初始化响应对象
	response := &ft.TBC20PoolPageResponse{
		PoolList:       make([]ft.TBC20PoolInfo, 0, len(result.Results)),
		TotalPoolCount: result.TotalCount,
	}

	// 如果没有找到任何池，返回空列表
	if len(result.Results) == 0 {
		log.InfoWithContextf(ctx, "未找到流动池信息，返回空列表")
		return response, nil
	}

	// 处理每个流动池信息
	for _, pool := range result.Results {
		// 初始化代币B的名称
		var tokenPairBName string

		// 只有当TokenContractId不为空时才查询代币信息
		if pool.TokenContractId != "" {
			// 获取代币B的信息
			tokenInfo, err := l.ftTokensDAO.GetFtTokenById(pool.TokenContractId)
			if err != nil {
				log.WarnWithContextf(ctx, "获取代币信息失败: 代币ID=%s, 错误=%v",
					pool.TokenContractId, err)
				// 出错时设置为空字符串，与Python代码行为一致
				tokenPairBName = ""
			} else if tokenInfo != nil {
				// 如果获取到代币信息，使用代币名称
				tokenPairBName = tokenInfo.FtName
			}
		}

		// 构建池信息
		poolInfo := ft.TBC20PoolInfo{
			PoolId:              pool.NftContractId,
			TokenPairAId:        "TBC",
			TokenPairAName:      "TBC",
			TokenPairBId:        pool.TokenContractId,
			TokenPairBName:      tokenPairBName,
			PoolCreateTimestamp: pool.CreateTimestamp,
		}

		// 添加到响应列表
		response.PoolList = append(response.PoolList, poolInfo)
	}

	log.InfoWithContextf(ctx, "WaitGroup并行查询完成: 总数=%d, 返回数量=%d",
		response.TotalPoolCount, len(response.PoolList))

	return response, nil
}
