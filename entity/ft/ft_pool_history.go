package ft

import (
	"fmt"
)

// TBC20PoolHistoryRequest 获取池子历史记录请求
type TBC20PoolHistoryRequest struct {
	PoolId string `uri:"pool_id" binding:"required"` // 池子ID
	Page   int    `uri:"page"`                       // 页码，从0开始
	Size   int    `uri:"size"`                       // 每页大小
}

// Validate 验证请求参数的合法性
func (req *TBC20PoolHistoryRequest) Validate() error {
	// 检查池子ID是否为空
	if req.PoolId == "" {
		return fmt.Errorf("池子ID不能为空")
	}

	// 检查页码是否合法
	if req.Page < 0 {
		return fmt.Errorf("页码不能小于0")
	}

	// 检查每页大小是否合法
	if req.Size <= 0 || req.Size > 100 {
		return fmt.Errorf("每页大小必须在1-100之间")
	}

	return nil
}

// TBC20PoolHistoryResponse 池子历史记录响应
type TBC20PoolHistoryResponse struct {
	Txid                        string `json:"txid"`                             // 交易ID
	PoolId                      string `json:"pool_id"`                          // 池子ID
	ExchangeAddress             string `json:"exchange_address"`                 // 交换地址
	FtLpBalanceChange           *int64 `json:"ft_lp_balance_change"`             // LP代币余额变化
	TokenPairAId                string `json:"token_pair_a_id"`                  // 代币A的ID（通常是"TBC"）
	TokenPairAName              string `json:"token_pair_a_name"`                // 代币A的名称（通常是"TBC"）
	TokenPairADecimal           int    `json:"token_pair_a_decimal"`             // 代币A的小数位数
	TokenPairAPoolBalanceChange *int64 `json:"token_pair_a_pool_balance_change"` // 代币A的池子余额变化
	TokenPairBId                string `json:"token_pair_b_id"`                  // 代币B的ID
	TokenPairBName              string `json:"token_pair_b_name"`                // 代币B的名称
	TokenPairBDecimal           int    `json:"token_pair_b_decimal"`             // 代币B的小数位数
	TokenPairBPoolBalanceChange *int64 `json:"token_pair_b_pool_balance_change"` // 代币B的池子余额变化
}
