package transaction_participants_dao

import (
	"context"
	"fmt"
	"ginproject/entity/dbtable"
	"ginproject/middleware/log"
	"ginproject/repo/db"
	"sync"
)

// AsyncResult 异步查询结果
type AsyncResult struct {
	Participants []*dbtable.TransactionParticipant
	Error        error
}

// 使用并发控制的全局变量
var (
	mutex          sync.Mutex
	activeQueries  int
	maxQueries     = 50 // 最大并发查询数
	querySemaphore = make(chan struct{}, maxQueries)
)

// 获取查询信号量
func acquireQueryPermission() {
	querySemaphore <- struct{}{}
	mutex.Lock()
	activeQueries++
	mutex.Unlock()
}

// 释放查询信号量
func releaseQueryPermission() {
	<-querySemaphore
	mutex.Lock()
	activeQueries--
	mutex.Unlock()
}

// GetParticipantsByTxHashes 根据交易哈希列表获取参与方信息
func GetParticipantsByTxHashes(ctx context.Context, txHashes []string) ([]*dbtable.TransactionParticipant, error) {
	if len(txHashes) == 0 {
		return []*dbtable.TransactionParticipant{}, nil
	}

	resultChan := make(chan *AsyncResult, 1)

	go func() {
		defer close(resultChan)

		// 获取查询权限
		acquireQueryPermission()
		defer releaseQueryPermission()

		log.InfoWithContext(ctx, "执行批量查询交易参与方", "txHash数量:", len(txHashes))

		var participants []*dbtable.TransactionParticipant
		result := db.GetDB().WithContext(ctx).
			Where("tx_hash IN ?", txHashes).
			Find(&participants)

		if result.Error != nil {
			log.ErrorWithContext(ctx, "批量查询交易参与方失败", "错误:", result.Error)
			resultChan <- &AsyncResult{
				Error: fmt.Errorf("批量查询交易参与方失败: %w", result.Error),
			}
			return
		}

		resultChan <- &AsyncResult{
			Participants: participants,
			Error:        nil,
		}
	}()

	// 等待并返回结果
	result := <-resultChan
	if result.Error != nil {
		return nil, result.Error
	}
	return result.Participants, nil
}

// GetParticipantsByTxHash 根据交易哈希获取参与方信息
func GetParticipantsByTxHash(ctx context.Context, txHash string) ([]*dbtable.TransactionParticipant, error) {
	resultChan := make(chan *AsyncResult, 1)

	go func() {
		defer close(resultChan)

		// 获取查询权限
		acquireQueryPermission()
		defer releaseQueryPermission()

		log.InfoWithContext(ctx, "执行查询交易参与方", "txHash:", txHash)

		var participants []*dbtable.TransactionParticipant
		result := db.GetDB().WithContext(ctx).
			Where("tx_hash = ?", txHash).
			Find(&participants)

		if result.Error != nil {
			log.ErrorWithContext(ctx, "查询交易参与方失败",
				"txHash:", txHash,
				"错误:", result.Error)
			resultChan <- &AsyncResult{
				Error: fmt.Errorf("查询交易参与方失败: %w", result.Error),
			}
			return
		}

		resultChan <- &AsyncResult{
			Participants: participants,
			Error:        nil,
		}
	}()

	// 等待并返回结果
	result := <-resultChan
	if result.Error != nil {
		return nil, result.Error
	}
	return result.Participants, nil
}

// GetParticipantsByTxHashAndRole 根据交易哈希和角色获取参与方信息
func GetParticipantsByTxHashAndRole(ctx context.Context, txHash string, role dbtable.Role) ([]*dbtable.TransactionParticipant, error) {
	resultChan := make(chan *AsyncResult, 1)

	go func() {
		defer close(resultChan)

		// 获取查询权限
		acquireQueryPermission()
		defer releaseQueryPermission()

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
			resultChan <- &AsyncResult{
				Error: fmt.Errorf("查询特定角色的交易参与方失败: %w", result.Error),
			}
			return
		}

		resultChan <- &AsyncResult{
			Participants: participants,
			Error:        nil,
		}
	}()

	// 等待并返回结果
	result := <-resultChan
	if result.Error != nil {
		return nil, result.Error
	}
	return result.Participants, nil
}
