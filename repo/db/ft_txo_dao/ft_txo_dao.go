package ft_txo_dao

import (
	"context"

	"ginproject/entity/dbtable"
	"ginproject/repo/db"

	"gorm.io/gorm"
)

// FtTxoDAO 用于管理ft_txo_set表操作的数据访问对象
type FtTxoDAO struct {
	db *gorm.DB
}

// NewFtTxoDAO 创建一个新的FtTxoDAO实例
func NewFtTxoDAO() *FtTxoDAO {
	return &FtTxoDAO{
		db: db.GetDB(),
	}
}

// InsertFtTxo 插入一条代币交易输出记录
func (dao *FtTxoDAO) InsertFtTxo(txo *dbtable.FtTxoSet) error {
	return dao.db.Create(txo).Error
}

// GetFtTxoByTxidVout 根据交易ID和输出索引获取代币交易输出
func (dao *FtTxoDAO) GetFtTxoByTxidVout(txid string, vout int) (*dbtable.FtTxoSet, error) {
	var txo dbtable.FtTxoSet
	err := dao.db.Where("utxo_txid = ? AND utxo_vout = ?", txid, vout).First(&txo).Error
	if err != nil {
		return nil, err
	}
	return &txo, nil
}

// UpdateFtTxo 更新代币交易输出信息
func (dao *FtTxoDAO) UpdateFtTxo(txo *dbtable.FtTxoSet) error {
	return dao.db.Save(txo).Error
}

// MarkFtTxoAsSpent 标记代币交易输出为已花费
func (dao *FtTxoDAO) MarkFtTxoAsSpent(txid string, vout int) error {
	return dao.db.Model(&dbtable.FtTxoSet{}).Where("utxo_txid = ? AND utxo_vout = ?", txid, vout).
		Update("if_spend", true).Error
}

// DeleteFtTxo 删除代币交易输出
func (dao *FtTxoDAO) DeleteFtTxo(txid string, vout int) error {
	return dao.db.Where("utxo_txid = ? AND utxo_vout = ?", txid, vout).Delete(&dbtable.FtTxoSet{}).Error
}

// GetFtTxosByHolderAndContract 根据持有者脚本和合约ID获取代币交易输出列表
func (dao *FtTxoDAO) GetFtTxosByHolderAndContract(holderScript string, contractId string) ([]*dbtable.FtTxoSet, error) {
	var txos []*dbtable.FtTxoSet
	err := dao.db.Where("ft_holder_combine_script = ? AND ft_contract_id = ?", holderScript, contractId).
		Find(&txos).Error
	return txos, err
}

// GetUnspentFtTxosByHolder 获取指定持有者的未花费代币交易输出
func (dao *FtTxoDAO) GetUnspentFtTxosByHolder(holderScript string) ([]*dbtable.FtTxoSet, error) {
	var txos []*dbtable.FtTxoSet
	err := dao.db.Where("ft_holder_combine_script = ? AND if_spend = ?", holderScript, false).
		Find(&txos).Error
	return txos, err
}

// GetUnspentFtTxosByHolderAndContract 获取指定持有者和合约的未花费代币交易输出
func (dao *FtTxoDAO) GetUnspentFtTxosByHolderAndContract(holderScript string, contractId string) ([]*dbtable.FtTxoSet, error) {
	var txos []*dbtable.FtTxoSet
	err := dao.db.Where("ft_holder_combine_script = ? AND ft_contract_id = ? AND if_spend = ?",
		holderScript, contractId, false).Find(&txos).Error
	return txos, err
}

// GetTotalBalanceByHolderAndContract 获取指定持有者和合约的未花费代币总余额
func (dao *FtTxoDAO) GetTotalBalanceByHolderAndContract(holderScript string, contractId string) (uint64, error) {
	type Result struct {
		TotalBalance uint64
	}
	var result Result
	err := dao.db.Model(&dbtable.FtTxoSet{}).Select("SUM(ft_balance) as total_balance").
		Where("ft_holder_combine_script = ? AND ft_contract_id = ? AND if_spend = ?",
			holderScript, contractId, false).Scan(&result).Error
	return result.TotalBalance, err
}

// BatchInsertFtTxos 批量插入代币交易输出记录
func (dao *FtTxoDAO) BatchInsertFtTxos(txos []*dbtable.FtTxoSet) error {
	return dao.db.CreateInBatches(txos, 100).Error
}

// GetTotalBalanceByHolder 获取指定持有者和合约的未花费代币总余额
func (dao *FtTxoDAO) GetTotalBalanceByHolder(ctx context.Context, holderScript string, contractId string) (uint64, error) {
	type Result struct {
		TotalBalance uint64
	}
	var result Result
	err := dao.db.Model(&dbtable.FtTxoSet{}).Select("SUM(ft_balance) as total_balance").
		Where("ft_holder_combine_script = ? AND ft_contract_id = ? AND if_spend = ?",
			holderScript, contractId, false).Scan(&result).Error

	// 如果没有记录，返回0
	if err == gorm.ErrRecordNotFound {
		return 0, nil
	}

	// 处理结果为nil的情况
	if result.TotalBalance == 0 {
		return 0, nil
	}

	return result.TotalBalance, err
}
