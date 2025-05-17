package nft_utxo_set_dao

import (
	"context"
	"fmt"
	"ginproject/entity/dbtable"
	"ginproject/middleware/log"
	"ginproject/repo/db"
	"sync"

	"gorm.io/gorm"
)

// NftUtxoSetDAO 用于管理nft_utxo_set表操作的数据访问对象
type NftUtxoSetDAO struct {
	db *gorm.DB
}

// NewNftUtxoSetDAO 创建一个新的NftUtxoSetDAO实例
func NewNftUtxoSetDAO() *NftUtxoSetDAO {
	return &NftUtxoSetDAO{
		db: db.GetDB(),
	}
}

// InsertNftUtxo 插入一条NFT UTXO记录
func (dao *NftUtxoSetDAO) InsertNftUtxo(utxo *dbtable.NftUtxoSet) error {
	return dao.db.Create(utxo).Error
}

// GetNftUtxoByContractId 根据合约ID获取NFT UTXO
func (dao *NftUtxoSetDAO) GetNftUtxoByContractId(contractId string) (*dbtable.NftUtxoSet, error) {
	var utxo dbtable.NftUtxoSet
	err := dao.db.Where("nft_contract_id = ?", contractId).First(&utxo).Error
	if err != nil {
		return nil, err
	}
	return &utxo, nil
}

// GetNftUtxoByUtxoId 根据UTXO ID获取NFT UTXO
func (dao *NftUtxoSetDAO) GetNftUtxoByUtxoId(utxoId string) (*dbtable.NftUtxoSet, error) {
	var utxo dbtable.NftUtxoSet
	err := dao.db.Where("nft_utxo_id = ?", utxoId).First(&utxo).Error
	if err != nil {
		return nil, err
	}
	return &utxo, nil
}

// UpdateNftUtxo 更新NFT UTXO信息
func (dao *NftUtxoSetDAO) UpdateNftUtxo(utxo *dbtable.NftUtxoSet) error {
	return dao.db.Save(utxo).Error
}

// DeleteNftUtxo 删除NFT UTXO
func (dao *NftUtxoSetDAO) DeleteNftUtxo(contractId string) error {
	return dao.db.Where("nft_contract_id = ?", contractId).Delete(&dbtable.NftUtxoSet{}).Error
}

// GetNftUtxosByCollection 根据集合ID获取NFT UTXO列表
func (dao *NftUtxoSetDAO) GetNftUtxosByCollection(collectionId string) ([]*dbtable.NftUtxoSet, error) {
	var utxos []*dbtable.NftUtxoSet
	err := dao.db.Where("collection_id = ?", collectionId).Find(&utxos).Error
	return utxos, err
}

// GetNftUtxosByHolder 根据持有者脚本哈希获取NFT UTXO列表
func (dao *NftUtxoSetDAO) GetNftUtxosByHolder(holderScriptHash string) ([]*dbtable.NftUtxoSet, error) {
	var utxos []*dbtable.NftUtxoSet
	err := dao.db.Where("nft_holder_script_hash = ?", holderScriptHash).Find(&utxos).Error
	return utxos, err
}

// GetNftUtxosWithPagination 分页获取NFT UTXO列表
func (dao *NftUtxoSetDAO) GetNftUtxosWithPagination(page, pageSize int) ([]*dbtable.NftUtxoSet, int64, error) {
	var utxos []*dbtable.NftUtxoSet
	var total int64

	// 获取总记录数
	if err := dao.db.Model(&dbtable.NftUtxoSet{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 获取分页数据
	offset := (page - 1) * pageSize
	if err := dao.db.Offset(offset).Limit(pageSize).Find(&utxos).Error; err != nil {
		return nil, 0, err
	}

	return utxos, total, nil
}

// GetPoolNftInfoByContractId 根据合约ID获取NFT池交易ID和余额
// 用于TBC20池NFT信息查询
func (dao *NftUtxoSetDAO) GetPoolNftInfoByContractId(ctx context.Context, ftContractId string) (string, uint64, error) {
	var result struct {
		NftUtxoId      string `gorm:"column:nft_utxo_id"`
		NftCodeBalance uint64 `gorm:"column:nft_code_balance"`
	}

	err := dao.db.WithContext(ctx).
		Table("TBC20721.nft_utxo_set").
		Select("nft_utxo_id, nft_code_balance").
		Where("nft_contract_id = ?", ftContractId).
		First(&result).Error

	if err != nil {
		return "", 0, err
	}

	return result.NftUtxoId, result.NftCodeBalance, nil
}

// GetPoolListByFtContractId 根据代币合约ID获取相关的流动池列表
func (dao *NftUtxoSetDAO) GetPoolListByFtContractId(ctx context.Context, ftContractId string) ([]struct {
	NftContractId   string `gorm:"column:nft_contract_id"`
	CreateTimestamp int64  `gorm:"column:nft_create_timestamp"`
}, error) {
	var results []struct {
		NftContractId   string `gorm:"column:nft_contract_id"`
		CreateTimestamp int64  `gorm:"column:nft_create_timestamp"`
	}

	// 从nft_utxo_set表中查询与指定代币相关的所有流动池
	// 使用nft_icon字段存储token_pair_a_id，并查询nft_holder_address='LP'的记录
	err := dao.db.WithContext(ctx).
		Table("TBC20721.nft_utxo_set").
		Select("nft_contract_id, nft_create_timestamp").
		Where("nft_holder_address = ? AND nft_icon = ?", "LP", ftContractId).
		Find(&results).Error

	return results, err
}

// GetAllPoolsWithPagination 异步分页获取所有流动池列表，使用并行查询优化性能
func (dao *NftUtxoSetDAO) GetAllPoolsWithPagination(ctx context.Context, page, size int) (<-chan struct {
	Results []struct {
		NftContractId   string `gorm:"column:nft_contract_id"`
		CreateTimestamp int64  `gorm:"column:nft_create_timestamp"`
		TokenContractId string `gorm:"column:nft_icon"`
	}
	TotalCount int64
	Error      error
}, error) {
	// 创建一个结果通道
	resultChan := make(chan struct {
		Results []struct {
			NftContractId   string `gorm:"column:nft_contract_id"`
			CreateTimestamp int64  `gorm:"column:nft_create_timestamp"`
			TokenContractId string `gorm:"column:nft_icon"`
		}
		TotalCount int64
		Error      error
	}, 1) // 缓冲为1，避免发送方阻塞

	// 启动异步协程执行并行查询
	go func() {
		defer close(resultChan) // 确保通道在函数结束时关闭

		// 使用日志记录异步查询开始
		log.InfoWithContextf(ctx, "开始并行异步查询流动池数据: 页码=%d, 每页大小=%d", page, size)

		// 查询结果结构
		var result struct {
			Results []struct {
				NftContractId   string `gorm:"column:nft_contract_id"`
				CreateTimestamp int64  `gorm:"column:nft_create_timestamp"`
				TokenContractId string `gorm:"column:nft_icon"`
			}
			TotalCount int64
			Error      error
		}

		// 创建两个通道，分别用于接收总数查询和数据查询的结果
		type countResult struct {
			Count int64
			Err   error
		}
		type dataResult struct {
			Data []struct {
				NftContractId   string `gorm:"column:nft_contract_id"`
				CreateTimestamp int64  `gorm:"column:nft_create_timestamp"`
				TokenContractId string `gorm:"column:nft_icon"`
			}
			Err error
		}

		countChan := make(chan countResult, 1)
		dataChan := make(chan dataResult, 1)

		// 启动协程1：查询总数
		go func() {
			var totalCount int64
			countErr := dao.db.WithContext(ctx).
				Table("TBC20721.nft_utxo_set").
				Where("nft_holder_address = ?", "LP").
				Count(&totalCount).Error

			countChan <- countResult{
				Count: totalCount,
				Err:   countErr,
			}
		}()

		// 启动协程2：查询分页数据
		go func() {
			var poolData []struct {
				NftContractId   string `gorm:"column:nft_contract_id"`
				CreateTimestamp int64  `gorm:"column:nft_create_timestamp"`
				TokenContractId string `gorm:"column:nft_icon"`
			}

			// 分页查询所有流动池
			offset := page * size // page从0开始
			err := dao.db.WithContext(ctx).
				Table("TBC20721.nft_utxo_set").
				Select("nft_contract_id, nft_create_timestamp, nft_icon").
				Where("nft_holder_address = ?", "LP").
				Offset(offset).
				Limit(size).
				Find(&poolData).Error

			dataChan <- dataResult{
				Data: poolData,
				Err:  err,
			}
		}()

		// 等待两个查询都完成
		cResult := <-countChan
		dResult := <-dataChan

		// 检查总数查询是否有错误
		if cResult.Err != nil {
			log.ErrorWithContextf(ctx, "并行异步查询流动池总数失败: %v", cResult.Err)
			result.Error = cResult.Err
			resultChan <- result
			return
		}

		// 检查数据查询是否有错误
		if dResult.Err != nil {
			log.ErrorWithContextf(ctx, "并行异步查询流动池数据失败: %v", dResult.Err)
			result.Error = dResult.Err
			resultChan <- result
			return
		}

		// 合并结果
		result.TotalCount = cResult.Count
		result.Results = dResult.Data

		log.InfoWithContextf(ctx, "并行异步查询流动池数据成功: 获取到%d条记录, 总数=%d",
			len(result.Results), result.TotalCount)
		resultChan <- result
	}()

	return resultChan, nil
}

// GetNftsByHolderWithPagination 根据持有者脚本哈希分页获取NFT列表
func (dao *NftUtxoSetDAO) GetNftsByHolderWithPagination(ctx context.Context, holderScriptHash string, page, size int) ([]*dbtable.NftUtxoSet, int64, error) {
	var nfts []*dbtable.NftUtxoSet
	var total int64

	// 计算起始索引
	offset := page * size

	// 获取总记录数
	if err := dao.db.WithContext(ctx).
		Model(&dbtable.NftUtxoSet{}).
		Where("nft_holder_script_hash = ?", holderScriptHash).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 获取分页数据，按照最后转移时间戳倒序排序
	if err := dao.db.WithContext(ctx).
		Where("nft_holder_script_hash = ?", holderScriptHash).
		Order("nft_last_transfer_timestamp DESC").
		Limit(size).
		Offset(offset).
		Find(&nfts).Error; err != nil {
		return nil, 0, err
	}

	return nfts, total, nil
}

// GetNftsByCollectionIdWithPagination 根据集合ID分页获取NFT列表
func (dao *NftUtxoSetDAO) GetNftsByCollectionIdWithPagination(ctx context.Context, collectionId string, page, size int) ([]*dbtable.NftUtxoSet, int64, error) {
	var nfts []*dbtable.NftUtxoSet
	var total int64

	// 计算起始索引
	offset := page * size

	// 获取总记录数
	if err := dao.db.WithContext(ctx).
		Model(&dbtable.NftUtxoSet{}).
		Where("collection_id = ?", collectionId).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 获取分页数据，按照集合索引排序
	if err := dao.db.WithContext(ctx).
		Where("collection_id = ?", collectionId).
		Order("collection_index").
		Limit(size).
		Offset(offset).
		Find(&nfts).Error; err != nil {
		return nil, 0, err
	}

	return nfts, total, nil
}

// GetNftsByContractIds 根据合约ID列表获取NFT信息
func (dao *NftUtxoSetDAO) GetNftsByContractIds(ctx context.Context, contractIds []string) ([]*dbtable.NftUtxoSet, error) {
	var nfts []*dbtable.NftUtxoSet

	if err := dao.db.WithContext(ctx).
		Where("nft_contract_id IN ?", contractIds).
		Find(&nfts).Error; err != nil {
		return nil, err
	}

	return nfts, nil
}

// GetNftByUtxoId 根据UTXO ID获取NFT信息
func (dao *NftUtxoSetDAO) GetNftByUtxoId(ctx context.Context, utxoId string) (*dbtable.NftUtxoSet, error) {
	var nft dbtable.NftUtxoSet
	err := dao.db.WithContext(ctx).
		Where("nft_utxo_id = ?", utxoId).
		First(&nft).Error
	if err != nil {
		return nil, err
	}
	return &nft, nil
}

// GetNftUtxoByContractIdWithContext 根据合约ID获取NFT UTXO（带上下文）
func (dao *NftUtxoSetDAO) GetNftUtxoByContractIdWithContext(ctx context.Context, contractId string) (*dbtable.NftUtxoSet, error) {
	var utxo dbtable.NftUtxoSet
	err := dao.db.WithContext(ctx).Where("nft_contract_id = ?", contractId).First(&utxo).Error
	if err != nil {
		return nil, err
	}
	return &utxo, nil
}

// GetNftsByCollectionAndIndex 根据集合ID和索引获取NFT列表
func (dao *NftUtxoSetDAO) GetNftsByCollectionAndIndex(ctx context.Context, collectionId string, collectionIndex int) ([]*dbtable.NftUtxoSet, error) {
	var nfts []*dbtable.NftUtxoSet
	err := dao.db.WithContext(ctx).
		Where("collection_id = ? AND collection_index = ?", collectionId, collectionIndex).
		Find(&nfts).Error
	return nfts, err
}

// GetAllPoolsWithPaginationParallel 使用WaitGroup实现并行查询的异步分页获取
func (dao *NftUtxoSetDAO) GetAllPoolsWithPaginationParallel(ctx context.Context, page, size int) (<-chan struct {
	Results []struct {
		NftContractId   string `gorm:"column:nft_contract_id"`
		CreateTimestamp int64  `gorm:"column:nft_create_timestamp"`
		TokenContractId string `gorm:"column:nft_icon"`
	}
	TotalCount int64
	Error      error
}, error) {
	// 创建结果通道
	resultChan := make(chan struct {
		Results []struct {
			NftContractId   string `gorm:"column:nft_contract_id"`
			CreateTimestamp int64  `gorm:"column:nft_create_timestamp"`
			TokenContractId string `gorm:"column:nft_icon"`
		}
		TotalCount int64
		Error      error
	}, 1)

	// 启动主协程
	go func() {
		defer close(resultChan)

		log.InfoWithContextf(ctx, "开始使用WaitGroup并行查询流动池数据: 页码=%d, 每页大小=%d", page, size)

		// 准备结果结构
		var result struct {
			Results []struct {
				NftContractId   string `gorm:"column:nft_contract_id"`
				CreateTimestamp int64  `gorm:"column:nft_create_timestamp"`
				TokenContractId string `gorm:"column:nft_icon"`
			}
			TotalCount int64
			Error      error
		}

		// 使用WaitGroup来等待两个查询都完成
		var wg sync.WaitGroup
		wg.Add(2) // 两个查询任务

		// 使用互斥锁保护共享的错误变量
		var mu sync.Mutex
		var firstError error

		// 协程1：查询总数
		go func() {
			defer wg.Done()

			var totalCount int64
			err := dao.db.WithContext(ctx).
				Table("TBC20721.nft_utxo_set").
				Where("nft_holder_address = ?", "LP").
				Count(&totalCount).Error

			if err != nil {
				mu.Lock()
				if firstError == nil {
					firstError = fmt.Errorf("查询总数失败: %w", err)
				}
				mu.Unlock()
				log.ErrorWithContextf(ctx, "WaitGroup并行查询总数失败: %v", err)
				return
			}

			mu.Lock()
			result.TotalCount = totalCount
			mu.Unlock()
		}()

		// 协程2：查询数据
		go func() {
			defer wg.Done()

			offset := page * size
			var poolData []struct {
				NftContractId   string `gorm:"column:nft_contract_id"`
				CreateTimestamp int64  `gorm:"column:nft_create_timestamp"`
				TokenContractId string `gorm:"column:nft_icon"`
			}

			err := dao.db.WithContext(ctx).
				Table("TBC20721.nft_utxo_set").
				Select("nft_contract_id, nft_create_timestamp, nft_icon").
				Where("nft_holder_address = ?", "LP").
				Offset(offset).
				Limit(size).
				Find(&poolData).Error

			if err != nil {
				mu.Lock()
				if firstError == nil {
					firstError = fmt.Errorf("查询数据失败: %w", err)
				}
				mu.Unlock()
				log.ErrorWithContextf(ctx, "WaitGroup并行查询数据失败: %v", err)
				return
			}

			mu.Lock()
			result.Results = poolData
			mu.Unlock()
		}()

		// 等待所有查询完成
		wg.Wait()

		// 检查是否有错误
		if firstError != nil {
			result.Error = firstError
			resultChan <- result
			return
		}

		log.InfoWithContextf(ctx, "WaitGroup并行查询流动池数据成功: 获取到%d条记录, 总数=%d",
			len(result.Results), result.TotalCount)
		resultChan <- result
	}()

	return resultChan, nil
}
