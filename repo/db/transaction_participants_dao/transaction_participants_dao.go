package transaction_participants_dao

import (
	"context"
	"fmt"
	"ginproject/entity/dbtable"
	"ginproject/middleware/log"
	"ginproject/repo/db"
)

// GetParticipantsByTxHashes 根据交易哈希列表获取参与方信息
func GetParticipantsByTxHashes(ctx context.Context, txHashes []string) ([]*dbtable.TransactionParticipant, error) {
	if len(txHashes) == 0 {
		return []*dbtable.TransactionParticipant{}, nil
	}

	log.InfoWithContext(ctx, "执行批量查询交易参与方", "txHash数量:", len(txHashes))

	var participants []*dbtable.TransactionParticipant
	result := db.GetDB().WithContext(ctx).
		Where("tx_hash IN ?", txHashes).
		Find(&participants)

	if result.Error != nil {
		log.ErrorWithContext(ctx, "批量查询交易参与方失败", "错误:", result.Error)
		return nil, fmt.Errorf("批量查询交易参与方失败: %w", result.Error)
	}

	return participants, nil
}

// GetParticipantsByTxHash 根据交易哈希获取参与方信息
func GetParticipantsByTxHash(ctx context.Context, txHash string) ([]*dbtable.TransactionParticipant, error) {
	log.InfoWithContext(ctx, "执行查询交易参与方", "txHash:", txHash)

	var participants []*dbtable.TransactionParticipant
	result := db.GetDB().WithContext(ctx).
		Where("tx_hash = ?", txHash).
		Find(&participants)

	if result.Error != nil {
		log.ErrorWithContext(ctx, "查询交易参与方失败",
			"txHash:", txHash,
			"错误:", result.Error)
		return nil, fmt.Errorf("查询交易参与方失败: %w", result.Error)
	}

	return participants, nil
}

// GetParticipantsByTxHashAndRole 根据交易哈希和角色获取参与方信息
func GetParticipantsByTxHashAndRole(ctx context.Context, txHash string, role dbtable.Role) ([]*dbtable.TransactionParticipant, error) {
	log.InfoWithContext(ctx, "执行查询特定角色的交易参与方",
		"txHash:", txHash,
		"role:", role)

	var participants []*dbtable.TransactionParticipant
	result := db.GetDB().WithContext(ctx).
		Where("tx_hash = ? AND role = ?", txHash, role).
		Find(&participants)

	if result.Error != nil {
		log.ErrorWithContext(ctx, "查询特定角色的交易参与方失败",
			"txHash:", txHash,
			"role:", role,
			"错误:", result.Error)
		return nil, fmt.Errorf("查询特定角色的交易参与方失败: %w", result.Error)
	}

	return participants, nil
}
