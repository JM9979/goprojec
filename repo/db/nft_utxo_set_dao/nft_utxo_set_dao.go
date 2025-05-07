package nft_utxo_set_dao

import (
	"context"
	"ginproject/entity/dbtable"
	"ginproject/repo/db"

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
		Select("nft_contract_id, nft_create_timestamp").
		Where("nft_holder_address = ? AND nft_icon = ?", "LP", ftContractId).
		Find(&results).Error

	return results, err
}

// GetAllPoolsWithPagination 分页获取所有流动池列表
func (dao *NftUtxoSetDAO) GetAllPoolsWithPagination(ctx context.Context, page, size int) ([]struct {
	NftContractId   string `gorm:"column:nft_contract_id"`
	CreateTimestamp int64  `gorm:"column:nft_create_timestamp"`
	TokenContractId string `gorm:"column:nft_icon"`
}, int64, error) {
	var results []struct {
		NftContractId   string `gorm:"column:nft_contract_id"`
		CreateTimestamp int64  `gorm:"column:nft_create_timestamp"`
		TokenContractId string `gorm:"column:nft_icon"`
	}

	// 计算总数量
	var totalCount int64
	countErr := dao.db.WithContext(ctx).
		Table("TBC20721.nft_utxo_set").
		Where("nft_holder_address = ?", "LP").
		Count(&totalCount).Error

	if countErr != nil {
		return nil, 0, countErr
	}

	// 分页查询所有流动池
	offset := page * size // page从0开始
	err := dao.db.WithContext(ctx).
		Table("TBC20721.nft_utxo_set").
		Select("nft_contract_id, nft_create_timestamp, nft_icon").
		Where("nft_holder_address = ?", "LP").
		Offset(offset).
		Limit(size).
		Find(&results).Error

	return results, totalCount, err
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
