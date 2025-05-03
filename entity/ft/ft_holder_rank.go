package ft

import (
	"fmt"
	"strconv"
)

// FtHolderRankRequest 获取代币持有者排名的请求参数
type FtHolderRankRequest struct {
	// 代币合约ID
	ContractId string `uri:"contract_id" binding:"required"`
	// 分页页码，从0开始
	Page int `uri:"page"`
	// 每页记录数
	Size int `uri:"size"`
}

// Validate 验证请求参数的合法性
func (req *FtHolderRankRequest) Validate() error {
	// 检查合约ID是否为空
	if req.ContractId == "" {
		return fmt.Errorf("合约ID不能为空")
	}

	// 检查页码是否合法
	if req.Page < 0 {
		return fmt.Errorf("页码必须大于等于0")
	}

	// 检查每页记录数是否合法
	if req.Size <= 0 || req.Size > 100 {
		return fmt.Errorf("每页记录数必须在1-100之间")
	}

	return nil
}

// ParsePage 解析页码，确保是有效数字
func (req *FtHolderRankRequest) ParsePage() (int, error) {
	return strconv.Atoi(strconv.Itoa(req.Page))
}

// ParseSize 解析每页记录数，确保是有效数字
func (req *FtHolderRankRequest) ParseSize() (int, error) {
	return strconv.Atoi(strconv.Itoa(req.Size))
}

// FtHolderRankResponse 获取代币持有者排名的响应
type FtHolderRankResponse struct {
	// 代币合约ID
	FtContractId string `json:"ft_contract_id"`
	// 代币精度
	FtDecimal int `json:"ft_decimal"`
	// 代币持有者总数
	FtHoldersCount int `json:"ft_holders_count"`
	// 持有者排名列表
	HolderRank []HolderRankInfo `json:"holder_rank"`
}

// HolderRankInfo 持有者排名信息
type HolderRankInfo struct {
	// 持有者地址
	Address string `json:"address"`
	// 持有代币余额
	Balance uint64 `json:"balance"`
	// 排名
	Rank int `json:"rank"`
	// 持有比例，取值0-1之间，保留4位小数
	HoldRatio float64 `json:"hold_ratio"`
}
