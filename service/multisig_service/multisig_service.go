package multisig_service

import (
	"net/http"

	"ginproject/entity/multisig"
	logic_multisig "ginproject/logic/multisig"
	"ginproject/middleware/log"

	"github.com/gin-gonic/gin"
)

// MultisigService 多签名服务接口
type MultisigService interface {
	GetMultiWalletByAddress(c *gin.Context)
}

// multisigService 多签名服务实现
type multisigService struct{}

// NewMultisigService 创建多签名服务实例
func NewMultisigService() MultisigService {
	return &multisigService{}
}

// GetMultiWalletByAddress 根据地址获取多签名钱包信息
func (s *multisigService) GetMultiWalletByAddress(c *gin.Context) {
	ctx := c.Request.Context()

	// 解析并验证参数
	var param multisig.AddressParam
	if err := c.ShouldBindUri(&param); err != nil {
		log.ErrorWithContext(ctx, "解析地址参数失败", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的地址参数"})
		return
	}

	// 验证地址参数
	if !param.IsValid() {
		log.ErrorWithContext(ctx, "地址参数无效", "address", param.Address)
		c.JSON(http.StatusBadRequest, gin.H{"error": "地址不能为空"})
		return
	}

	log.InfoWithContext(ctx, "开始获取多签名钱包信息", "address", param.Address)

	// 调用逻辑层获取多签名钱包信息
	result, err := logic_multisig.QueryMultiWalletByAddress(ctx, param.Address)
	if err != nil {
		log.ErrorWithContext(ctx, "获取多签名钱包信息失败", "address", param.Address, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取多签名钱包信息失败"})
		return
	}

	log.InfoWithContext(ctx, "成功获取多签名钱包信息",
		"address", param.Address,
		"count", len(result.MultiWalletList))

	// 返回结果
	c.JSON(http.StatusOK, result)
}
