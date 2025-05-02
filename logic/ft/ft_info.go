package ft

import (
	"context"
	"fmt"
	"math"

	"ginproject/entity/ft"
	"ginproject/middleware/log"
)

// GetFtInfoByContractId 根据合约ID获取FT信息
func (l *FtLogic) GetFtInfoByContractId(ctx context.Context, req *ft.FtInfoContractIdRequest) (*ft.TBC20FTInfoResponse, error) {
	// 使用entity层的验证逻辑
	if err := req.Validate(); err != nil {
		log.ErrorWithContextf(ctx, "参数验证失败: %v", err)
		return nil, fmt.Errorf("参数验证失败: %v", err)
	}

	// 获取代币信息
	ftToken, err := l.ftTokensDAO.GetFtTokenById(req.ContractId)
	if err != nil {
		log.ErrorWithContextf(ctx, "获取代币信息失败: %v", err)
		return nil, fmt.Errorf("获取代币信息失败: %v", err)
	}

	// 计算考虑小数位后的供应量
	ftSupply := float64(ftToken.FtSupply) / math.Pow10(int(ftToken.FtDecimal))

	// 构造响应
	response := &ft.TBC20FTInfoResponse{
		FtContractId:           ftToken.FtContractId,
		FtCodeScript:           ftToken.FtCodeScript,
		FtTapeScript:           ftToken.FtTapeScript,
		FtSupply:               ftSupply,
		FtDecimal:              int(ftToken.FtDecimal),
		FtName:                 ftToken.FtName,
		FtSymbol:               ftToken.FtSymbol,
		FtDescription:          ftToken.FtDescription,
		FtOriginUtxo:           ftToken.FtOriginUtxo,
		FtCreatorCombineScript: ftToken.FtCreatorCombineScript,
		FtHoldersCount:         ftToken.FtHoldersCount,
		FtIconUrl:              ftToken.FtIconUrl,
		FtCreateTimestamp:      ftToken.FtCreateTimestamp,
		FtTokenPrice:           ftToken.FtTokenPrice,
	}

	log.InfoWithContextf(ctx, "FT信息查询成功: 合约ID=%s", req.ContractId)

	return response, nil
}
