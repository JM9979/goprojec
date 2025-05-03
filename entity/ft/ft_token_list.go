package ft

import (
	"fmt"
	"strings"

	"ginproject/entity/utility"
)

// FtTokenListRequest 代币列表请求参数
type FtTokenListRequest struct {
	Page    int    `uri:"page"`
	Size    int    `uri:"size"`
	OrderBy string `uri:"order_by" binding:"required"`
}

// FtTokenInfo 代币信息
type FtTokenInfo struct {
	FtContractId      string  `json:"ftContractId"`
	FtSupply          float64 `json:"ftSupply"`
	FtDecimal         int     `json:"ftDecimal"`
	FtName            string  `json:"ftName"`
	FtSymbol          string  `json:"ftSymbol"`
	FtDescription     string  `json:"ftDescription"`
	FtCreatorAddress  string  `json:"ftCreatorAddress"`
	FtCreateTimestamp int     `json:"ftCreateTimestamp"`
	FtTokenPrice      float64 `json:"ftTokenPrice"`
	FtHoldersCount    int     `json:"ftHoldersCount"`
	FtIconUrl         string  `json:"ftIconUrl"`
}

// FtTokenListData 代币列表数据
type FtTokenListData struct {
	FtTokenCount int            `json:"ftTokenCount"`
	FtTokenList  []*FtTokenInfo `json:"ftTokenList"`
}

// NewFtTokenListResponse 创建成功的代币列表响应
func NewFtTokenListResponse(count int, tokens []*FtTokenInfo) utility.APIResponse {
	data := FtTokenListData{
		FtTokenCount: count,
		FtTokenList:  tokens,
	}

	return utility.NewSuccessResponse(data)
}

// Validate 验证请求参数是否合法
func (req *FtTokenListRequest) Validate() error {
	// 验证排序字段是否合法
	validOrderFields := []string{"ftCreateTimestamp", "ftHoldersCount"}
	isValid := false
	for _, field := range validOrderFields {
		if strings.EqualFold(req.OrderBy, field) {
			isValid = true
			break
		}
	}

	if req.Page < 0 {
		return fmt.Errorf("页码不能小于0")
	}

	if req.Size < 0 {
		return fmt.Errorf("每页大小不能小于0")
	}

	if !isValid {
		return fmt.Errorf("无效的排序字段，只支持'ftCreateTimestamp'或'ftHoldersCount'")
	}

	return nil
}
