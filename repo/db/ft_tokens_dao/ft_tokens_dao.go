package ft_tokens_dao

import (
	"context"

	"ginproject/entity/dbtable"
	"ginproject/middleware/log"
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

// GetFtCodeScript 根据合约ID获取代币代码脚本
func (dao *FtTokensDAO) GetFtCodeScript(ctx context.Context, contractId string) (string, error) {
	var token dbtable.FtTokens

	// 查询代币代码脚本
	err := dao.db.Where("ft_contract_id = ?", contractId).Select("ft_code_script").First(&token).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			log.WarnWithContextf(ctx, "未找到合约ID对应的代币信息: %s", contractId)
			return "", nil
		}
		log.ErrorWithContextf(ctx, "查询代币代码脚本失败: %v", err)
		return "", err
	}

	log.InfoWithContextf(ctx, "成功获取合约ID[%s]的代币代码脚本", contractId)
	return token.FtCodeScript, nil
}

// GetFtCodeScriptAndDecimal 根据合约ID获取代币代码脚本和精度
func (dao *FtTokensDAO) GetFtCodeScriptAndDecimal(ctx context.Context, contractId string) (string, uint8, error) {
	var token dbtable.FtTokens

	// 查询代币代码脚本和精度
	err := dao.db.Where("ft_contract_id = ?", contractId).Select("ft_code_script, ft_decimal").First(&token).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			log.WarnWithContextf(ctx, "未找到合约ID对应的代币信息: %s", contractId)
			return "", 0, nil
		}
		log.ErrorWithContextf(ctx, "查询代币代码脚本和精度失败: %v", err)
		return "", 0, err
	}

	log.InfoWithContextf(ctx, "成功获取合约ID[%s]的代币代码脚本和精度", contractId)
	return token.FtCodeScript, token.FtDecimal, nil
}

// GetTokensPageByCreateTime 根据创建时间排序获取代币分页列表
func (dao *FtTokensDAO) GetTokensPageByCreateTime(ctx context.Context, page, size int) ([]*dbtable.FtTokens, int64, error) {
	var tokens []*dbtable.FtTokens
	var total int64

	// 获取总记录数
	if err := dao.db.Model(&dbtable.FtTokens{}).Count(&total).Error; err != nil {
		log.ErrorWithContextf(ctx, "获取代币总数失败: %v", err)
		return nil, 0, err
	}

	// 获取分页数据，按创建时间排序
	offset := page * size
	if err := dao.db.Order("ft_create_timestamp DESC").
		Offset(offset).
		Limit(size).
		Find(&tokens).Error; err != nil {
		log.ErrorWithContextf(ctx, "获取代币分页列表失败: %v", err)
		return nil, 0, err
	}

	log.InfoWithContextf(ctx, "成功获取代币分页列表，总数: %d, 当前页: %d, 每页大小: %d",
		total, page, size)
	return tokens, total, nil
}

// GetTokensPageByHoldersCount 根据持有人数量排序获取代币分页列表
func (dao *FtTokensDAO) GetTokensPageByHoldersCount(ctx context.Context, page, size int) ([]*dbtable.FtTokens, int64, error) {
	var tokens []*dbtable.FtTokens
	var total int64

	// 获取总记录数
	if err := dao.db.Model(&dbtable.FtTokens{}).Count(&total).Error; err != nil {
		log.ErrorWithContextf(ctx, "获取代币总数失败: %v", err)
		return nil, 0, err
	}

	// 获取分页数据，按持有人数量排序
	offset := page * size
	if err := dao.db.Order("ft_holders_count DESC").
		Offset(offset).
		Limit(size).
		Find(&tokens).Error; err != nil {
		log.ErrorWithContextf(ctx, "获取代币分页列表失败: %v", err)
		return nil, 0, err
	}

	log.InfoWithContextf(ctx, "成功获取代币分页列表，总数: %d, 当前页: %d, 每页大小: %d",
		total, page, size)
	return tokens, total, nil
}
