package transactions_dao

import (
	"context"
	"fmt"
	"ginproject/entity/dbtable"
	"ginproject/middleware/log"
	"ginproject/repo/db"
)

// GetTransactionByTxHash 根据交易哈希获取交易信息
func GetTransactionByTxHash(ctx context.Context, txHash string) (*dbtable.Transaction, error) {
	log.InfoWithContext(ctx, "执行查询交易信息", "txHash:", txHash)
	var transaction dbtable.Transaction
	result := db.GetDB().WithContext(ctx).Where("tx_hash = ?", txHash).First(&transaction)

	if result.Error != nil {
		log.ErrorWithContext(ctx, "查询交易信息失败", "txHash:", txHash, "错误:", result.Error)
		return nil, fmt.Errorf("查询交易信息失败: %w", result.Error)
	}

	return &transaction, nil
}

// GetTransactionsByTxHashes 根据交易哈希列表批量获取交易信息
func GetTransactionsByTxHashes(ctx context.Context, txHashes []string) ([]*dbtable.Transaction, error) {
	if len(txHashes) == 0 {
		return []*dbtable.Transaction{}, nil
	}

	log.InfoWithContext(ctx, "执行批量查询交易信息", "txHash数量:", len(txHashes))

	var transactions []*dbtable.Transaction
	result := db.GetDB().WithContext(ctx).Where("tx_hash IN ?", txHashes).Find(&transactions)

	if result.Error != nil {
		log.ErrorWithContext(ctx, "批量查询交易信息失败", "错误:", result.Error)
		return nil, fmt.Errorf("批量查询交易信息失败: %w", result.Error)
	}

	return transactions, nil
}

// CountTransactions 计算交易总数
func CountTransactions(ctx context.Context) (int64, error) {
	log.InfoWithContext(ctx, "执行计算交易总数")

	var count int64
	result := db.GetDB().WithContext(ctx).Model(&dbtable.Transaction{}).Count(&count)

	if result.Error != nil {
		log.ErrorWithContext(ctx, "计算交易总数失败", "错误:", result.Error)
		return 0, fmt.Errorf("计算交易总数失败: %w", result.Error)
	}

	return count, nil
}
