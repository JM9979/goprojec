package ft_balance_dao

import (
	"context"

	"ginproject/entity/dbtable"
	"ginproject/repo/db"

	"gorm.io/gorm"
)

// FtBalanceDAO 用于管理ft_balance表操作的数据访问对象
type FtBalanceDAO struct {
	db *gorm.DB
}

// NewFtBalanceDAO 创建一个新的FtBalanceDAO实例
func NewFtBalanceDAO() *FtBalanceDAO {
	return &FtBalanceDAO{
		db: db.GetDB(),
	}
}

// InsertFtBalance 插入一条代币余额记录
func (dao *FtBalanceDAO) InsertFtBalance(balance *dbtable.FtBalance) error {
	return dao.db.Create(balance).Error
}

// GetFtBalance 根据持有者脚本和合约ID获取代币余额
func (dao *FtBalanceDAO) GetFtBalance(holderScript string, contractId string) (*dbtable.FtBalance, error) {
	var balance dbtable.FtBalance
	err := dao.db.Where("ft_holder_combine_script = ? AND ft_contract_id = ?", holderScript, contractId).First(&balance).Error
	if err != nil {
		return nil, err
	}
	return &balance, nil
}

// UpdateFtBalance 更新代币余额信息
func (dao *FtBalanceDAO) UpdateFtBalance(balance *dbtable.FtBalance) error {
	return dao.db.Save(balance).Error
}

// DeleteFtBalance 删除代币余额
func (dao *FtBalanceDAO) DeleteFtBalance(holderScript string, contractId string) error {
	return dao.db.Where("ft_holder_combine_script = ? AND ft_contract_id = ?", holderScript, contractId).Delete(&dbtable.FtBalance{}).Error
}

// GetFtBalancesByHolder 获取持有者的所有代币余额
func (dao *FtBalanceDAO) GetFtBalancesByHolder(holderScript string) ([]*dbtable.FtBalance, error) {
	var balances []*dbtable.FtBalance
	err := dao.db.Where("ft_holder_combine_script = ?", holderScript).Find(&balances).Error
	return balances, err
}

// GetFtBalancesByContractId 获取某代币的所有持有者余额
func (dao *FtBalanceDAO) GetFtBalancesByContractId(contractId string) ([]*dbtable.FtBalance, error) {
	var balances []*dbtable.FtBalance
	err := dao.db.Where("ft_contract_id = ?", contractId).Find(&balances).Error
	return balances, err
}

// GetFtBalancesWithPagination 分页获取代币余额列表
func (dao *FtBalanceDAO) GetFtBalancesWithPagination(page, pageSize int) ([]*dbtable.FtBalance, int64, error) {
	var balances []*dbtable.FtBalance
	var total int64

	// 获取总记录数
	if err := dao.db.Model(&dbtable.FtBalance{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 获取分页数据
	offset := (page - 1) * pageSize
	if err := dao.db.Offset(offset).Limit(pageSize).Find(&balances).Error; err != nil {
		return nil, 0, err
	}

	return balances, total, nil
}

// GetSumBalanceByContractId 获取某代币的总余额
func (dao *FtBalanceDAO) GetSumBalanceByContractId(contractId string) (uint64, error) {
	type Result struct {
		TotalBalance uint64
	}
	var result Result
	err := dao.db.Model(&dbtable.FtBalance{}).Select("SUM(ft_balance) as total_balance").
		Where("ft_contract_id = ?", contractId).Scan(&result).Error
	return result.TotalBalance, err
}

// GetHoldersCountByContractId 获取某代币的持有者数量
func (dao *FtBalanceDAO) GetHoldersCountByContractId(contractId string) (int64, error) {
	var count int64
	err := dao.db.Model(&dbtable.FtBalance{}).Where("ft_contract_id = ?", contractId).Count(&count).Error
	return count, err
}

// GetFtBalanceRankByContractId 获取代币持有者排名列表
func (dao *FtBalanceDAO) GetFtBalanceRankByContractId(ctx context.Context, contractId string, page, size int) ([]*dbtable.FtBalance, error) {
	var balances []*dbtable.FtBalance

	// 计算偏移量
	offset := page * size

	// 查询持有者排名，按持有余额降序排序
	err := dao.db.Where("ft_contract_id = ?", contractId).
		Order("ft_balance DESC").
		Offset(offset).
		Limit(size).
		Find(&balances).Error

	if err != nil {
		return nil, err
	}

	return balances, nil
}
