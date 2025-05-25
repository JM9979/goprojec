package ft

import (
	"fmt"
)

// TBC20PoolNFTInfoRequest 获取NFT池信息请求
type TBC20PoolNFTInfoRequest struct {
	FtContractId string `uri:"ft_contract_id" binding:"required"`
}

// Validate 验证请求参数的合法性
func (req *TBC20PoolNFTInfoRequest) Validate() error {
	// 检查合约ID是否为空
	if req.FtContractId == "" {
		return fmt.Errorf("合约ID不能为空")
	}

	// 可以增加更多的验证逻辑，例如检查合约ID格式
	if len(req.FtContractId) < 8 {
		return fmt.Errorf("合约ID格式不正确")
	}

	return nil
}

// TBC20PoolNFTInfoResponse 获取NFT池信息响应
type TBC20PoolNFTInfoResponse struct {
	FtLpBalance           *int64  `json:"ft_lp_balance,omitempty"`
	FtABalance            *int64  `json:"ft_a_balance,omitempty"`
	TbcBalance            *int64  `json:"tbc_balance,omitempty"`
	PoolVersion           *int64  `json:"pool_version,omitempty"`
	PoolServiceFeeRate    *int    `json:"pool_service_fee_rate,omitempty"`
	PoolServiceProvider   *string `json:"pool_service_provider,omitempty"`
	FtLpPartialHash       *string `json:"ft_lp_partial_hash,omitempty"`
	FtAPartialHash        *string `json:"ft_a_partial_hash,omitempty"`
	FtAContractTxid       *string `json:"ft_a_contract_txid,omitempty"`
	PoolNftCodeScript     *string `json:"pool_nft_code_script,omitempty"`
	CurrentPoolNftTxid    *string `json:"current_pool_nft_txid,omitempty"`
	CurrentPoolNftVout    *int64  `json:"current_pool_nft_vout,omitempty"`
	CurrentPoolNftBalance *int64  `json:"current_pool_nft_balance,omitempty"`
}

// ErrorResponse 错误响应
type ErrorResponse struct {
	Error string `json:"error"`
}
