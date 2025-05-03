package ft

import (
	"ginproject/repo/db/ft_balance_dao"
	"ginproject/repo/db/ft_tokens_dao"
	"ginproject/repo/db/ft_txo_dao"
	"ginproject/repo/db/nft_utxo_set_dao"
)

// FtLogic 代表FT代币相关的业务逻辑
type FtLogic struct {
	ftTokensDAO  *ft_tokens_dao.FtTokensDAO
	ftTxoDAO     *ft_txo_dao.FtTxoDAO
	ftBalanceDAO *ft_balance_dao.FtBalanceDAO
	ftPoolNftDAO *nft_utxo_set_dao.NftUtxoSetDAO
}

// NewFtLogic 创建一个新的FtLogic实例
func NewFtLogic() *FtLogic {
	return &FtLogic{
		ftTokensDAO:  ft_tokens_dao.NewFtTokensDAO(),
		ftTxoDAO:     ft_txo_dao.NewFtTxoDAO(),
		ftBalanceDAO: ft_balance_dao.NewFtBalanceDAO(),
		ftPoolNftDAO: nft_utxo_set_dao.NewNftUtxoSetDAO(),
	}
}
