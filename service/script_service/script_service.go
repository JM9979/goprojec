package script_service

import (
	"net/http"

	"ginproject/entity/script"
	"ginproject/middleware/log"
	"ginproject/repo/rpc/electrumx"

	"github.com/gin-gonic/gin"
)

// ScriptService 脚本服务
type ScriptService struct{}

// NewScriptService 创建新的脚本服务实例
func NewScriptService() *ScriptService {
	return &ScriptService{}
}

// GetScriptUnspent 获取脚本的未花费交易输出
func (s *ScriptService) GetScriptUnspent(c *gin.Context) {
	// 获取上下文和脚本哈希参数
	ctx := c.Request.Context()
	scriptHash := c.Param("script_hash")

	// 参数校验
	if err := script.ValidateScriptHash(scriptHash); err != nil {
		log.ErrorWithContext(ctx, "脚本哈希参数无效",
			"scriptHash", scriptHash,
			"error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 记录API调用
	log.InfoWithContext(ctx, "开始获取脚本未花费交易输出",
		"scriptHash", scriptHash)

	// 调用RPC获取未花费交易输出
	utxos, err := electrumx.GetScriptHashUnspent(ctx, scriptHash)
	if err != nil {
		log.ErrorWithContext(ctx, "获取脚本未花费交易输出失败",
			"scriptHash", scriptHash,
			"error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取脚本未花费交易输出失败"})
		return
	}

	// 返回结果
	log.InfoWithContext(ctx, "成功获取脚本未花费交易输出",
		"scriptHash", scriptHash,
		"count", len(utxos))
	c.JSON(http.StatusOK, utxos)
}

// GetScriptHistory 获取脚本的历史记录
func (s *ScriptService) GetScriptHistory(c *gin.Context) {
	// 获取上下文和脚本哈希参数
	ctx := c.Request.Context()
	scriptHash := c.Param("script_hash")

	// 参数校验
	if err := script.ValidateScriptHash(scriptHash); err != nil {
		log.ErrorWithContext(ctx, "脚本哈希参数无效",
			"scriptHash", scriptHash,
			"error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 记录API调用
	log.InfoWithContext(ctx, "开始获取脚本历史记录",
		"scriptHash", scriptHash)

	// 调用RPC获取脚本历史记录
	history, err := electrumx.GetScriptHashHistory(ctx, scriptHash)
	if err != nil {
		log.ErrorWithContext(ctx, "获取脚本历史记录失败",
			"scriptHash", scriptHash,
			"error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取脚本历史记录失败"})
		return
	}

	// 返回结果
	log.InfoWithContext(ctx, "成功获取脚本历史记录",
		"scriptHash", scriptHash,
		"count", len(history))
	c.JSON(http.StatusOK, history)
}
