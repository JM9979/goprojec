package transactions_dao

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
	Transaction  *dbtable.Transaction
	Transactions []*dbtable.Transaction
	Count        int64
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

// GetTransactionByTxHash 根据交易哈希获取交易信息
func GetTransactionByTxHash(ctx context.Context, txHash string) (*dbtable.Transaction, error) {
	resultChan := make(chan *AsyncResult, 1)

	go func() {
		defer close(resultChan)

		// 获取查询权限
		acquireQueryPermission()
		defer releaseQueryPermission()

		log.InfoWithContext(ctx, "执行查询交易信息", "txHash:", txHash)
		var transaction dbtable.Transaction
		result := db.GetDB().WithContext(ctx).Where("tx_hash = ?", txHash).First(&transaction)

		if result.Error != nil {
			log.ErrorWithContext(ctx, "查询交易信息失败", "txHash:", txHash, "错误:", result.Error)
			resultChan <- &AsyncResult{
				Error: fmt.Errorf("查询交易信息失败: %w", result.Error),
			}
			return
		}

		resultChan <- &AsyncResult{
			Transaction: &transaction,
			Error:       nil,
		}
	}()

	// 等待并返回结果
	result := <-resultChan
	if result.Error != nil {
		return nil, result.Error
	}
	return result.Transaction, nil
}

// GetTransactionsByTxHashes 根据交易哈希列表批量获取交易信息
func GetTransactionsByTxHashes(ctx context.Context, txHashes []string) ([]*dbtable.Transaction, error) {
	if len(txHashes) == 0 {
		return []*dbtable.Transaction{}, nil
	}

	resultChan := make(chan *AsyncResult, 1)

	go func() {
		defer close(resultChan)

		// 获取查询权限
		acquireQueryPermission()
		defer releaseQueryPermission()

		log.InfoWithContext(ctx, "执行批量查询交易信息", "txHash数量:", len(txHashes))

		var transactions []*dbtable.Transaction
		result := db.GetDB().WithContext(ctx).Where("tx_hash IN ?", txHashes).Find(&transactions)

		if result.Error != nil {
			log.ErrorWithContext(ctx, "批量查询交易信息失败", "错误:", result.Error)
			resultChan <- &AsyncResult{
				Error: fmt.Errorf("批量查询交易信息失败: %w", result.Error),
			}
			return
		}

		resultChan <- &AsyncResult{
			Transactions: transactions,
			Error:        nil,
		}
	}()

	// 等待并返回结果
	result := <-resultChan
	if result.Error != nil {
		return nil, result.Error
	}
	return result.Transactions, nil
}

// CountTransactions 计算交易总数
func CountTransactions(ctx context.Context) (int64, error) {
	resultChan := make(chan *AsyncResult, 1)

	go func() {
		defer close(resultChan)

		// 获取查询权限
		acquireQueryPermission()
		defer releaseQueryPermission()

		log.InfoWithContext(ctx, "执行计算交易总数")

		var count int64
		result := db.GetDB().WithContext(ctx).Model(&dbtable.Transaction{}).Count(&count)

		if result.Error != nil {
			log.ErrorWithContext(ctx, "计算交易总数失败", "错误:", result.Error)
			resultChan <- &AsyncResult{
				Error: fmt.Errorf("计算交易总数失败: %w", result.Error),
			}
			return
		}

		resultChan <- &AsyncResult{
			Count: count,
			Error: nil,
		}
	}()

	// 等待并返回结果
	result := <-resultChan
	if result.Error != nil {
		return 0, result.Error
	}
	return result.Count, nil
}
