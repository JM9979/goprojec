package address_transactions_dao

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
	Transactions []*dbtable.AddressTransaction
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

// GetAddressTransactions 根据地址获取相关交易记录
func GetAddressTransactions(ctx context.Context, address string, offset, limit int) ([]*dbtable.AddressTransaction, error) {
	resultChan := make(chan *AsyncResult, 1)

	go func() {
		defer close(resultChan)

		// 获取查询权限
		acquireQueryPermission()
		defer releaseQueryPermission()

		log.InfoWithContext(ctx, "执行查询地址交易记录",
			"address:", address,
			"offset:", offset,
			"limit:", limit)

		var transactions []*dbtable.AddressTransaction
		result := db.GetDB().WithContext(ctx).
			Where("address = ?", address).
			Order("Fid DESC").
			Offset(offset).
			Limit(limit).
			Find(&transactions)

		if result.Error != nil {
			log.ErrorWithContext(ctx, "查询地址交易记录失败",
				"address:", address,
				"错误:", result.Error)
			resultChan <- &AsyncResult{
				Error: fmt.Errorf("查询地址交易记录失败: %w", result.Error),
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

// CountAddressTransactions 统计地址的交易记录数量
func CountAddressTransactions(ctx context.Context, address string) (int64, error) {
	resultChan := make(chan *AsyncResult, 1)

	go func() {
		defer close(resultChan)

		// 获取查询权限
		acquireQueryPermission()
		defer releaseQueryPermission()

		log.InfoWithContext(ctx, "执行统计地址交易数量", "address:", address)

		var count int64
		result := db.GetDB().WithContext(ctx).
			Model(&dbtable.AddressTransaction{}).
			Where("address = ?", address).
			Count(&count)

		if result.Error != nil {
			log.ErrorWithContext(ctx, "统计地址交易数量失败",
				"address:", address,
				"错误:", result.Error)
			resultChan <- &AsyncResult{
				Error: fmt.Errorf("统计地址交易数量失败: %w", result.Error),
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

// GetAddressTransactionsByTxHashes 根据交易哈希列表获取地址相关的交易记录
func GetAddressTransactionsByTxHashes(ctx context.Context, address string, txHashes []string) ([]*dbtable.AddressTransaction, error) {
	if len(txHashes) == 0 {
		return []*dbtable.AddressTransaction{}, nil
	}

	resultChan := make(chan *AsyncResult, 1)

	go func() {
		defer close(resultChan)

		// 获取查询权限
		acquireQueryPermission()
		defer releaseQueryPermission()

		log.InfoWithContext(ctx, "执行批量查询地址交易记录",
			"address:", address,
			"txHash数量:", len(txHashes))

		var transactions []*dbtable.AddressTransaction
		result := db.GetDB().WithContext(ctx).
			Where("address = ? AND tx_hash IN ?", address, txHashes).
			Find(&transactions)

		if result.Error != nil {
			log.ErrorWithContext(ctx, "批量查询地址交易记录失败",
				"address:", address,
				"错误:", result.Error)
			resultChan <- &AsyncResult{
				Error: fmt.Errorf("批量查询地址交易记录失败: %w", result.Error),
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
