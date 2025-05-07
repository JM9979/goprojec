package nft_collections_dao

import (
	"context"

	"ginproject/entity/dbtable"
	"ginproject/repo/db"

	"gorm.io/gorm"
)

// NftCollectionsDAO 用于管理nft_collections表操作的数据访问对象
type NftCollectionsDAO struct {
	db *gorm.DB
}

// NewNftCollectionsDAO 创建一个新的NftCollectionsDAO实例
func NewNftCollectionsDAO() *NftCollectionsDAO {
	return &NftCollectionsDAO{
		db: db.GetDB(),
	}
}

// InsertNftCollection 插入一条NFT集合记录
func (dao *NftCollectionsDAO) InsertNftCollection(collection *dbtable.NftCollections) error {
	return dao.db.Create(collection).Error
}

// GetNftCollectionById 根据集合ID获取NFT集合
func (dao *NftCollectionsDAO) GetNftCollectionById(collectionId string) (*dbtable.NftCollections, error) {
	var collection dbtable.NftCollections
	err := dao.db.Where("collection_id = ?", collectionId).First(&collection).Error
	if err != nil {
		return nil, err
	}
	return &collection, nil
}

// UpdateNftCollection 更新NFT集合信息
func (dao *NftCollectionsDAO) UpdateNftCollection(collection *dbtable.NftCollections) error {
	return dao.db.Save(collection).Error
}

// DeleteNftCollection 删除NFT集合
func (dao *NftCollectionsDAO) DeleteNftCollection(collectionId string) error {
	return dao.db.Where("collection_id = ?", collectionId).Delete(&dbtable.NftCollections{}).Error
}

// GetCollectionsByCreator 根据创建者脚本哈希获取集合列表
func (dao *NftCollectionsDAO) GetCollectionsByCreator(creatorScriptHash string) ([]*dbtable.NftCollections, error) {
	var collections []*dbtable.NftCollections
	err := dao.db.Where("collection_creator_script_hash = ?", creatorScriptHash).Find(&collections).Error
	return collections, err
}

// GetCollectionsWithPagination 分页获取NFT集合列表
func (dao *NftCollectionsDAO) GetCollectionsWithPagination(page, pageSize int) ([]*dbtable.NftCollections, int64, error) {
	var collections []*dbtable.NftCollections
	var total int64

	// 获取总记录数
	if err := dao.db.Model(&dbtable.NftCollections{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 获取分页数据
	offset := (page - 1) * pageSize
	if err := dao.db.Offset(offset).Limit(pageSize).Find(&collections).Error; err != nil {
		return nil, 0, err
	}

	return collections, total, nil
}

// GetCollectionsByAddressWithPagination 根据创建者地址分页获取集合列表
func (dao *NftCollectionsDAO) GetCollectionsByAddressWithPagination(ctx context.Context, address string, page, size int) ([]*dbtable.NftCollections, int64, error) {
	var collections []*dbtable.NftCollections
	var total int64

	// 计算起始索引
	offset := page * size

	// 获取总记录数
	if err := dao.db.WithContext(ctx).
		Model(&dbtable.NftCollections{}).
		Where("collection_creator_address = ?", address).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 获取分页数据，按照创建时间倒序排序
	if err := dao.db.WithContext(ctx).
		Where("collection_creator_address = ?", address).
		Order("collection_create_timestamp DESC").
		Limit(size).
		Offset(offset).
		Find(&collections).Error; err != nil {
		return nil, 0, err
	}

	return collections, total, nil
}

// GetAllCollectionsWithPagination 获取所有集合并分页
func (dao *NftCollectionsDAO) GetAllCollectionsWithPagination(ctx context.Context, page, size int) ([]*dbtable.NftCollections, int64, error) {
	var collections []*dbtable.NftCollections
	var total int64

	// 计算起始索引
	offset := page * size

	// 获取总记录数
	if err := dao.db.WithContext(ctx).
		Model(&dbtable.NftCollections{}).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 获取分页数据，按照创建时间倒序排序
	if err := dao.db.WithContext(ctx).
		Order("collection_create_timestamp DESC").
		Limit(size).
		Offset(offset).
		Find(&collections).Error; err != nil {
		return nil, 0, err
	}

	return collections, total, nil
}

// GetDetailCollectionInfo 获取集合详细信息
func (dao *NftCollectionsDAO) GetDetailCollectionInfo(ctx context.Context, collectionId string) (*dbtable.NftCollections, error) {
	var collection dbtable.NftCollections
	err := dao.db.WithContext(ctx).
		Where("collection_id = ?", collectionId).
		First(&collection).Error
	if err != nil {
		return nil, err
	}
	return &collection, nil
}

// GetCollectionIconAndDescription 获取集合图标和描述
func (dao *NftCollectionsDAO) GetCollectionIconAndDescription(ctx context.Context, collectionId string) (string, string, error) {
	var result struct {
		CollectionIcon        string `gorm:"column:collection_icon"`
		CollectionDescription string `gorm:"column:collection_description"`
	}

	err := dao.db.WithContext(ctx).
		Model(&dbtable.NftCollections{}).
		Select("collection_icon, collection_description").
		Where("collection_id = ?", collectionId).
		First(&result).Error
	if err != nil {
		return "", "", err
	}

	return result.CollectionIcon, result.CollectionDescription, nil
}
