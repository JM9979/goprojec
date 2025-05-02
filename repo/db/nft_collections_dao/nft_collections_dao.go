package nft_collections_dao

import (
	"ginproject/repo/db"
	"ginproject/entity/dbtable"

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
