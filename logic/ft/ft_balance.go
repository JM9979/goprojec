package ft

import (
	"context"
	"fmt"

	"ginproject/entity/ft"
	"ginproject/middleware/log"
	"ginproject/repo/db/ft_tokens_dao"
	"ginproject/repo/db/ft_txo_dao"
	"ginproject/repo/db/nft_utxo_set_dao"
)

// FtLogic FT代币逻辑处理
type FtLogic struct {
	ftTxoDAO     *ft_txo_dao.FtTxoDAO
	ftTokensDAO  *ft_tokens_dao.FtTokensDAO
	ftPoolNftDAO *nft_utxo_set_dao.NftUtxoSetDAO
}

// NewFtLogic 创建FtLogic实例
func NewFtLogic() *FtLogic {
	return &FtLogic{
		ftTxoDAO:     ft_txo_dao.NewFtTxoDAO(),
		ftTokensDAO:  ft_tokens_dao.NewFtTokensDAO(),
		ftPoolNftDAO: nft_utxo_set_dao.NewNftUtxoSetDAO(),
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

// GetMultiFtBalanceByAddress 获取地址持有的多个代币余额
func (l *FtLogic) GetMultiFtBalanceByAddress(ctx context.Context, req *ft.FtBalanceMultiContractRequest) ([]ft.TBC20FTBalanceResponse, error) {
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

	// 初始化响应结果切片
	responseList := make([]ft.TBC20FTBalanceResponse, 0, len(req.FtContractId))

	// 遍历每个合约ID，查询余额
	for _, contractId := range req.FtContractId {
		// 获取代币小数位数
		ftDecimal, err := l.ftTokensDAO.GetFtDecimalByContractId(ctx, contractId)
		if err != nil {
			log.ErrorWithContextf(ctx, "获取代币小数位数失败，合约ID=%s: %v", contractId, err)
			// 跳过错误的合约，继续处理其他合约
			continue
		}

		// 调用DAO层获取余额
		ftBalance, err := l.ftTxoDAO.GetTotalBalanceByHolder(ctx, combineScript, contractId)
		if err != nil {
			log.ErrorWithContextf(ctx, "查询FT余额失败，合约ID=%s: %v", contractId, err)
			// 跳过错误的合约，继续处理其他合约
			continue
		}

		// 构造单个代币的余额响应
		response := ft.TBC20FTBalanceResponse{
			CombineScript: combineScript,
			FtContractId:  contractId,
			FtDecimal:     int(ftDecimal),
			FtBalance:     ftBalance,
		}

		// 添加到响应列表
		responseList = append(responseList, response)
	}

	log.InfoWithContextf(ctx, "批量查询FT余额成功，合约数量: %d", len(responseList))
	return responseList, nil
}

// GetTokensListHeldByAddress 获取地址持有的所有代币列表
func (l *FtLogic) GetTokensListHeldByAddress(ctx context.Context, req *ft.TBC20TokenListHeldByAddressRequest) (*ft.TBC20TokenListHeldByAddressResponse, error) {
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

	log.InfoWithContextf(ctx, "查询地址[%s]持有的代币列表, 组合脚本: %s", req.Address, combineScript)

	// 获取该地址持有的所有代币合约ID
	contractIds, err := l.ftTxoDAO.GetFtContractIdsByHolder(ctx, combineScript)
	if err != nil {
		log.ErrorWithContextf(ctx, "获取地址持有的代币合约ID列表失败: %v", err)
		return nil, fmt.Errorf("获取地址持有的代币合约ID列表失败: %v", err)
	}

	if len(contractIds) == 0 {
		log.InfoWithContextf(ctx, "地址[%s]未持有任何代币", req.Address)
		return &ft.TBC20TokenListHeldByAddressResponse{
			Address:    req.Address,
			TokenCount: 0,
			TokenList:  []ft.TokenInfo{},
		}, nil
	}

	log.InfoWithContextf(ctx, "地址[%s]持有的代币数量: %d", req.Address, len(contractIds))

	// 初始化代币列表
	tokenList := make([]ft.TokenInfo, 0, len(contractIds))

	// 遍历所有合约ID，查询详细信息
	for _, contractId := range contractIds {
		// 查询代币详细信息
		token, err := l.ftTokensDAO.GetFtTokenById(contractId)
		if err != nil {
			log.WarnWithContextf(ctx, "获取代币[%s]详细信息失败: %v，跳过此代币", contractId, err)
			continue
		}

		// 跳过LP代币（流动性代币）
		if token.FtSymbol == "LP" || token.FtName == "LP Token" {
			log.InfoWithContextf(ctx, "跳过LP代币: %s", contractId)
			continue
		}

		// 获取代币余额
		ftBalance, err := l.ftTxoDAO.GetTotalBalanceByHolder(ctx, combineScript, contractId)
		if err != nil {
			log.WarnWithContextf(ctx, "获取代币[%s]余额失败: %v，跳过此代币", contractId, err)
			continue
		}

		// 如果余额为0，则跳过
		if ftBalance == 0 {
			log.InfoWithContextf(ctx, "跳过余额为0的代币: %s", contractId)
			continue
		}

		// 添加代币信息到列表
		tokenInfo := ft.TokenInfo{
			FtContractId: contractId,
			FtDecimal:    int(token.FtDecimal),
			FtBalance:    ftBalance,
			FtName:       token.FtName,
			FtSymbol:     token.FtSymbol,
		}
		tokenList = append(tokenList, tokenInfo)
	}

	// 构造响应
	response := &ft.TBC20TokenListHeldByAddressResponse{
		Address:    req.Address,
		TokenCount: len(tokenList),
		TokenList:  tokenList,
	}

	log.InfoWithContextf(ctx, "成功获取地址[%s]持有的代币列表，数量: %d", req.Address, len(tokenList))
	return response, nil
}
