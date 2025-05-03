package ft_txo_dao

import (
	"context"
	"fmt"

	"ginproject/entity/dbtable"
	"ginproject/entity/ft"
	"ginproject/middleware/log"
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

// GetFtUtxoInfo 根据交易ID和输出索引获取代币余额、持有者组合脚本和合约ID
func (dao *FtTxoDAO) GetFtUtxoInfo(ctx context.Context, txid string, vout int) (uint64, string, string, error) {
	var result struct {
		FtBalance             uint64
		FtHolderCombineScript string
		FtContractId          string
	}

	err := dao.db.Model(&dbtable.FtTxoSet{}).
		Select("ft_balance, ft_holder_combine_script, ft_contract_id").
		Where("utxo_txid = ? AND utxo_vout = ?", txid, vout).
		First(&result).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return 0, "", "", nil
		}
		return 0, "", "", err
	}

	return result.FtBalance, result.FtHolderCombineScript, result.FtContractId, nil
}

// GetPoolHistoryByPoolId 根据池子ID查询历史记录
func (dao *FtTxoDAO) GetPoolHistoryByPoolId(ctx context.Context, poolId string, page int, size int) ([]ft.TBC20PoolHistoryResponse, error) {
	log.InfoWithContextf(ctx, "查询池子历史记录: 池子ID=%s, 页码=%d, 每页大小=%d", poolId, page, size)

	var results []ft.TBC20PoolHistoryResponse

	// 使用联表查询获取更完整的池子历史记录信息
	type Result struct {
		UtxoTxid              string
		FtHolderCombineScript string
		FtContractId          string
		FtBalance             uint64
		UtxoBalance           uint64
		FtName                string
		FtDecimal             int
	}

	var queryResults []Result

	// 联表查询ft_txo_set和ft_tokens表，获取代币名称和精度
	err := dao.db.Table("TBC20721.ft_txo_set as t1").
		Select("t1.utxo_txid, t1.ft_holder_combine_script, t1.ft_contract_id, t1.ft_balance, t1.utxo_balance, t2.ft_name, t2.ft_decimal").
		Joins("left join TBC20721.ft_tokens as t2 on t1.ft_contract_id = t2.ft_contract_id").
		Where("t1.utxo_txid = ? OR t1.ft_holder_combine_script = ?", poolId, poolId).
		Order("t1.id DESC").
		Offset(page * size).
		Limit(size).
		Scan(&queryResults).Error

	if err != nil {
		log.ErrorWithContextf(ctx, "查询池子历史记录失败: %v", err)
		return nil, err
	}

	// 转换为响应类型
	for _, result := range queryResults {
		// 根据实际逻辑填充池子历史记录
		history := ft.TBC20PoolHistoryResponse{
			Txid:                        result.UtxoTxid,
			PoolId:                      poolId,
			ExchangeAddress:             result.FtHolderCombineScript,
			FtLpBalanceChange:           formatBalanceChange(int64(result.FtBalance / 2)), // LP代币通常是总量的一半
			TokenPairAId:                "TBC",
			TokenPairAName:              "TBC",
			TokenPairADecimal:           6, // TBC默认精度
			TokenPairAPoolBalanceChange: formatBalanceChange(int64(result.UtxoBalance)),
			TokenPairBId:                result.FtContractId,
			TokenPairBName:              result.FtName,
			TokenPairBDecimal:           result.FtDecimal,
			TokenPairBPoolBalanceChange: formatBalanceChange(int64(result.FtBalance)),
		}

		results = append(results, history)
	}

	log.InfoWithContextf(ctx, "查询池子历史记录成功: 池子ID=%s, 记录数=%d", poolId, len(results))
	return results, nil
}

// formatBalanceChange 格式化余额变化为字符串，添加正负号
func formatBalanceChange(balance int64) string {
	if balance > 0 {
		return "+" + fmt.Sprintf("%d", balance)
	}
	return fmt.Sprintf("%d", balance)
}
