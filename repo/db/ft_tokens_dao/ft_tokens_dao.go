package ft_tokens_dao

import (
	"context"

	"ginproject/entity/dbtable"
	"ginproject/repo/db"

	"gorm.io/gorm"
)

// FtTokensDAO 用于管理ft_tokens表操作的数据访问对象
type FtTokensDAO struct {
	db *gorm.DB
}

// NewFtTokensDAO 创建一个新的FtTokensDAO实例
func NewFtTokensDAO() *FtTokensDAO {
	return &FtTokensDAO{
		db: db.GetDB(),
	}
}

// InsertFtToken 插入一条代币记录
func (dao *FtTokensDAO) InsertFtToken(token *dbtable.FtTokens) error {
	return dao.db.Create(token).Error
}

// GetFtTokenById 根据合约ID获取代币
func (dao *FtTokensDAO) GetFtTokenById(contractId string) (*dbtable.FtTokens, error) {
	var token dbtable.FtTokens
	err := dao.db.Where("ft_contract_id = ?", contractId).First(&token).Error
	if err != nil {
		return nil, err
	}
	return &token, nil
}

// GetFtTokenByOriginUtxo 根据源UTXO获取代币
func (dao *FtTokensDAO) GetFtTokenByOriginUtxo(originUtxo string) (*dbtable.FtTokens, error) {
	var token dbtable.FtTokens
	err := dao.db.Where("ft_origin_utxo = ?", originUtxo).First(&token).Error
	if err != nil {
		return nil, err
	}
	return &token, nil
}

// UpdateFtToken 更新代币信息
func (dao *FtTokensDAO) UpdateFtToken(token *dbtable.FtTokens) error {
	return dao.db.Save(token).Error
}

// DeleteFtToken 删除代币
func (dao *FtTokensDAO) DeleteFtToken(contractId string) error {
	return dao.db.Where("ft_contract_id = ?", contractId).Delete(&dbtable.FtTokens{}).Error
}

// GetFtTokensByName 根据名称查询代币列表
func (dao *FtTokensDAO) GetFtTokensByName(name string) ([]*dbtable.FtTokens, error) {
	var tokens []*dbtable.FtTokens
	err := dao.db.Where("ft_name LIKE ?", "%"+name+"%").Find(&tokens).Error
	return tokens, err
}

// GetFtTokensBySymbol 根据符号查询代币列表
func (dao *FtTokensDAO) GetFtTokensBySymbol(symbol string) ([]*dbtable.FtTokens, error) {
	var tokens []*dbtable.FtTokens
	err := dao.db.Where("ft_symbol LIKE ?", "%"+symbol+"%").Find(&tokens).Error
	return tokens, err
}

// GetFtTokensByCreator 根据创建者查询代币列表
func (dao *FtTokensDAO) GetFtTokensByCreator(creatorCombineScript string) ([]*dbtable.FtTokens, error) {
	var tokens []*dbtable.FtTokens
	err := dao.db.Where("ft_creator_combine_script = ?", creatorCombineScript).Find(&tokens).Error
	return tokens, err
}

// GetFtTokensWithPagination 分页获取代币列表
func (dao *FtTokensDAO) GetFtTokensWithPagination(page, pageSize int) ([]*dbtable.FtTokens, int64, error) {
	var tokens []*dbtable.FtTokens
	var total int64

	// 获取总记录数
	if err := dao.db.Model(&dbtable.FtTokens{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 获取分页数据
	offset := (page - 1) * pageSize
	if err := dao.db.Offset(offset).Limit(pageSize).Find(&tokens).Error; err != nil {
		return nil, 0, err
	}

	return tokens, total, nil
}

// GetFtDecimalByContractId 根据合约ID获取代币小数位数
func (dao *FtTokensDAO) GetFtDecimalByContractId(ctx context.Context, contractId string) (uint8, error) {
	var token dbtable.FtTokens
	err := dao.db.Where("ft_contract_id = ?", contractId).Select("ft_decimal").First(&token).Error
	if err != nil {
		return 0, err
	}
	return token.FtDecimal, nil
}
