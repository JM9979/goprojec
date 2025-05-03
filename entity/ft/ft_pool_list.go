package ft

import (
	"fmt"
)

// TBC20PoolListRequest 获取代币相关的流动池列表请求
type TBC20PoolListRequest struct {
	FtContractId string `uri:"ft_contract_id" binding:"required"`
}

// Validate 验证请求参数的合法性
func (req *TBC20PoolListRequest) Validate() error {
	// 检查合约ID是否为空
	if req.FtContractId == "" {
		return fmt.Errorf("合约ID不能为空")
	}

	// 检查合约ID格式
	if len(req.FtContractId) < 8 {
		return fmt.Errorf("合约ID格式不正确")
	}

	return nil
}

// TBC20PoolInfo 流动池信息
type TBC20PoolInfo struct {
	PoolId              string `json:"pool_id"`               // 池ID
	TokenPairAId        string `json:"token_pair_a_id"`       // 代币A的ID
	TokenPairAName      string `json:"token_pair_a_name"`     // 代币A的名称
	TokenPairBId        string `json:"token_pair_b_id"`       // 代币B的ID
	TokenPairBName      string `json:"token_pair_b_name"`     // 代币B的名称
	PoolCreateTimestamp int64  `json:"pool_create_timestamp"` // 池创建时间戳
}

// TBC20PoolListResponse 获取代币相关的流动池列表响应
type TBC20PoolListResponse struct {
	PoolList       []TBC20PoolInfo `json:"pool_list"`        // 流动池列表
	TotalPoolCount int64           `json:"total_pool_count"` // 总数量
}
