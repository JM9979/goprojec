package ft_txo_dao

import (
	"context"
	"errors"
	"fmt"

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

// Create 创建新的交易输出记录
func (d *FtTxoDAO) Create(ctx context.Context, txo *dbtable.FtTxoSet) error {
	// 添加记录到数据库
	if err := d.db.WithContext(ctx).Create(txo).Error; err != nil {
		return fmt.Errorf("创建交易输出记录失败: %w", err)
	}
	return nil
}

// GetByTxidAndVout 根据交易ID和输出索引获取记录
func (d *FtTxoDAO) GetByTxidAndVout(ctx context.Context, txid string, vout int) (*dbtable.FtTxoSet, error) {
	var txo dbtable.FtTxoSet
	if err := d.db.WithContext(ctx).Where("utxo_txid = ? AND utxo_vout = ?", txid, vout).First(&txo).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // 返回nil表示没有找到记录
		}
		return nil, fmt.Errorf("查询交易输出记录失败: %w", err)
	}
	return &txo, nil
}

// ListByHolderAndContract 根据持有者和合约ID列出交易输出记录
func (d *FtTxoDAO) ListByHolderAndContract(ctx context.Context, holderScript string, contractId string) ([]*dbtable.FtTxoSet, error) {
	var txos []*dbtable.FtTxoSet
	query := d.db.WithContext(ctx).Where("ft_holder_combine_script = ? AND ft_contract_id = ?", holderScript, contractId)
	if err := query.Find(&txos).Error; err != nil {
		return nil, fmt.Errorf("查询持有者交易输出记录失败: %w", err)
	}
	return txos, nil
}

// ListUnspentByHolder 获取持有者的未花费交易输出记录
func (d *FtTxoDAO) ListUnspentByHolder(ctx context.Context, holderScript string) ([]*dbtable.FtTxoSet, error) {
	var txos []*dbtable.FtTxoSet
	query := d.db.WithContext(ctx).Where("ft_holder_combine_script = ? AND if_spend = ?", holderScript, false)
	if err := query.Find(&txos).Error; err != nil {
		return nil, fmt.Errorf("查询未花费交易输出记录失败: %w", err)
	}
	return txos, nil
}

// UpdateSpentStatus 更新交易输出记录的花费状态
func (d *FtTxoDAO) UpdateSpentStatus(ctx context.Context, txid string, vout int, isSpent bool) error {
	result := d.db.WithContext(ctx).Model(&dbtable.FtTxoSet{}).
		Where("utxo_txid = ? AND utxo_vout = ?", txid, vout).
		Update("if_spend", isSpent)

	if result.Error != nil {
		return fmt.Errorf("更新交易输出花费状态失败: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("未找到指定的交易输出记录")
	}

	return nil
}

// GetTotalBalanceByHolder 获取持有者的总余额
func (d *FtTxoDAO) GetTotalBalanceByHolder(ctx context.Context, holderScript string, contractId string) (uint64, error) {
	var total uint64

	query := d.db.WithContext(ctx).
		Model(&dbtable.FtTxoSet{}).
		Select("COALESCE(SUM(ft_balance), 0) as total").
		Where("ft_holder_combine_script = ? AND ft_contract_id = ? AND if_spend = ?", holderScript, contractId, false)

	if err := query.Scan(&total).Error; err != nil {
		return 0, fmt.Errorf("计算总余额失败: %w", err)
	}

	return total, nil
}

// BatchCreate 批量创建交易输出记录
func (d *FtTxoDAO) BatchCreate(ctx context.Context, txos []*dbtable.FtTxoSet) error {
	if len(txos) == 0 {
		return nil
	}

	if err := d.db.WithContext(ctx).CreateInBatches(txos, 100).Error; err != nil {
		return fmt.Errorf("批量创建交易输出记录失败: %w", err)
	}

	return nil
}

// DeleteByTxidAndVout 删除指定的交易输出记录
func (d *FtTxoDAO) DeleteByTxidAndVout(ctx context.Context, txid string, vout int) error {
	result := d.db.WithContext(ctx).
		Where("utxo_txid = ? AND utxo_vout = ?", txid, vout).
		Delete(&dbtable.FtTxoSet{})

	if result.Error != nil {
		return fmt.Errorf("删除交易输出记录失败: %w", result.Error)
	}

	return nil
}
