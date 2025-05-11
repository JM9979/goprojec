package multisig

// MultiWalletResponse 多签名钱包响应
type MultiWalletResponse struct {
	MultiWalletList []MultiWallet `json:"multi_wallet_list"`
}

// MultiWallet 多签名钱包信息
type MultiWallet struct {
	MultiAddress string   `json:"multi_address"`
	PubkeyList   []string `json:"pubkey_list"`
}

// AddressParam 地址请求参数
type AddressParam struct {
	Address string `uri:"address" binding:"required"`
}

// IsValid 检查地址参数是否有效
func (p *AddressParam) IsValid() bool {
	return len(p.Address) > 0
}
