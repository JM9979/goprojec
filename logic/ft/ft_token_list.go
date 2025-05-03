package ft

import (
	"context"
	"fmt"
	"strings"

	"ginproject/entity/dbtable"
	"ginproject/entity/ft"
	"ginproject/entity/utility"
	"ginproject/middleware/log"
	"ginproject/repo/db/ft_tokens_dao"
)

// GetFtTokenList 获取代币列表
func (l *FtLogic) GetFtTokenList(ctx context.Context, req *ft.FtTokenListRequest) (*ft.FtTokenListData, error) {
	// 验证请求参数
	if err := req.Validate(); err != nil {
		log.ErrorWithContextf(ctx, "获取代币列表参数验证失败: %v", err)
		return nil, err
	}

	// 查询代币列表
	tokens, total, err := l.queryTokens(ctx, req)
	if err != nil {
		log.ErrorWithContextf(ctx, "获取代币列表失败: %v", err)
		return nil, err
	}

	// 转换为响应格式
	tokenInfoList := l.convertTokensToInfoList(ctx, tokens)

	// 创建响应
	response := &ft.FtTokenListData{
		FtTokenCount: int(total),
		FtTokenList:  tokenInfoList,
	}
	log.InfoWithContextf(ctx, "成功获取代币列表，总数: %d, 当前页: %d, 每页大小: %d",
		total, req.Page, req.Size)

	return response, nil
}

// queryTokens 根据请求参数查询代币列表
func (l *FtLogic) queryTokens(ctx context.Context, req *ft.FtTokenListRequest) ([]*dbtable.FtTokens, int64, error) {
	var tokens []*dbtable.FtTokens
	var total int64
	var err error

	// 创建DAO
	dao := ft_tokens_dao.NewFtTokensDAO()

	// 根据排序字段选择查询方法
	if strings.EqualFold(req.OrderBy, "ftCreateTimestamp") {
		tokens, total, err = dao.GetTokensPageByCreateTime(ctx, req.Page, req.Size)
	} else if strings.EqualFold(req.OrderBy, "ftHoldersCount") {
		tokens, total, err = dao.GetTokensPageByHoldersCount(ctx, req.Page, req.Size)
	}

	return tokens, total, err
}

// convertTokensToInfoList 将数据库实体转换为API响应
func (l *FtLogic) convertTokensToInfoList(ctx context.Context, tokens []*dbtable.FtTokens) []*ft.FtTokenInfo {
	tokenInfoList := make([]*ft.FtTokenInfo, 0, len(tokens))

	for _, token := range tokens {
		// 将代币创建者脚本转换为地址
		creatorAddress, err := utility.ConvertCombineScriptToAddress(token.FtCreatorCombineScript)
		if err != nil {
			log.WarnWithContextf(ctx, "转换代币创建者地址失败: %v, 使用原始脚本", err)
			creatorAddress = token.FtCreatorCombineScript
		}

		// 计算供应量的浮点数表示（总供应量/10^精度）
		divisor := float64(1)
		for i := 0; i < int(token.FtDecimal); i++ {
			divisor *= 10
		}
		supply := float64(token.FtSupply) / divisor

		tokenInfo := &ft.FtTokenInfo{
			FtContractId:      token.FtContractId,
			FtSupply:          supply,
			FtDecimal:         int(token.FtDecimal),
			FtName:            token.FtName,
			FtSymbol:          token.FtSymbol,
			FtDescription:     token.FtDescription,
			FtCreatorAddress:  creatorAddress,
			FtCreateTimestamp: token.FtCreateTimestamp,
			FtTokenPrice:      fmt.Sprintf("%f", token.FtTokenPrice),
			FtHoldersCount:    token.FtHoldersCount,
			FtIconUrl:         token.FtIconUrl,
		}
		tokenInfoList = append(tokenInfoList, tokenInfo)
	}

	return tokenInfoList
}
