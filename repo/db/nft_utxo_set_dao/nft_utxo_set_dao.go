package nft_utxo_set_dao

import (
	"ginproject/repo/db"
	"ginproject/entity/dbtable"

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
