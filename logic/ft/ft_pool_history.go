package ft

import (
	"context"
	"fmt"
	"strings"

	"ginproject/entity/blockchain"
	"ginproject/entity/ft"
	"ginproject/entity/utility"
	"ginproject/middleware/log"
	"ginproject/repo/db/ft_tokens_dao"
	rpcBlockchain "ginproject/repo/rpc/blockchain"
	"ginproject/repo/rpc/electrumx"
)

// GetPoolHistory 获取指定池的历史交易记录
func (l *FtLogic) GetPoolHistory(ctx context.Context, req *ft.TBC20PoolHistoryRequest) ([]ft.TBC20PoolHistoryResponse, error) {
	// 使用entity层的验证逻辑
	if err := req.Validate(); err != nil {
		log.ErrorWithContextf(ctx, "参数验证失败: %v", err)
		return nil, fmt.Errorf("参数验证失败: %w", err)
	}

	// 创建返回结果切片
	poolHistoryList := make([]ft.TBC20PoolHistoryResponse, 0)

	// 记录请求开始日志
	log.InfoWithContextf(ctx, "开始获取池历史记录: 池ID=%s, 页码=%d, 页大小=%d",
		req.PoolId, req.Page, req.Size)

	// 从区块链获取池交易信息
	decodeTxResult := <-rpcBlockchain.DecodeTxHash(ctx, req.PoolId)
	if decodeTxResult.Error != nil {
		log.ErrorWithContextf(ctx, "获取池交易信息失败: %v", decodeTxResult.Error)
		return nil, fmt.Errorf("获取池交易信息失败: %w", decodeTxResult.Error)
	}

	// 类型断言获取解码交易结果
	decodeTx, ok := decodeTxResult.Result.(*blockchain.TransactionResponse)
	if !ok || decodeTx == nil {
		log.ErrorWithContextf(ctx, "解析池交易结果失败: 解码结果类型错误")
		return nil, fmt.Errorf("解析池交易结果失败: 解码结果类型错误")
	}

	// 确保有输出
	if len(decodeTx.Vout) == 0 {
		log.ErrorWithContextf(ctx, "池交易没有输出: %s", req.PoolId)
		return nil, fmt.Errorf("池交易没有输出")
	}

	// 获取池脚本哈希
	scriptPubKeyHex := decodeTx.Vout[0].ScriptPubKey.Hex
	scriptHash, err := utility.ConvertStrToSha256(scriptPubKeyHex)
	if err != nil {
		log.ErrorWithContextf(ctx, "转换脚本哈希失败: %v", err)
		return nil, fmt.Errorf("转换脚本哈希失败: %w", err)
	}

	// 获取池历史
	scriptHistory, err := electrumx.GetScriptHashHistory(ctx, scriptHash)
	if err != nil {
		log.ErrorWithContextf(ctx, "获取池历史失败: %v", err)
		return nil, fmt.Errorf("获取池历史失败: %w", err)
	}

	// 反转历史列表以按时间降序排序
	for i, j := 0, len(scriptHistory)-1; i < j; i, j = i+1, j-1 {
		scriptHistory[i], scriptHistory[j] = scriptHistory[j], scriptHistory[i]
	}

	// 应用分页
	startIndex := req.Page * req.Size
	endIndex := startIndex + req.Size
	if startIndex >= len(scriptHistory) {
		log.InfoWithContextf(ctx, "请求的页码超出范围: 页码=%d, 总记录数=%d", req.Page, len(scriptHistory))
		return poolHistoryList, nil
	}
	if endIndex > len(scriptHistory) {
		endIndex = len(scriptHistory)
	}
	pageHistory := scriptHistory[startIndex:endIndex]

	// 处理每条历史记录
	for _, historyItem := range pageHistory {
		// 获取交易哈希
		txHash := historyItem.TxHash

		// 创建历史记录响应
		historyResponse := ft.TBC20PoolHistoryResponse{
			Txid:              txHash,
			PoolId:            req.PoolId,
			TokenPairAId:      "TBC",
			TokenPairAName:    "TBC",
			TokenPairADecimal: 6,
		}

		// 获取历史交易详情
		historyDecodeTxResult := <-rpcBlockchain.DecodeTxHash(ctx, txHash)
		if historyDecodeTxResult.Error != nil {
			log.WarnWithContextf(ctx, "获取历史交易详情失败, 跳过: %v", historyDecodeTxResult.Error)
			continue
		}

		// 类型断言获取历史交易
		historyDecodeTx, ok := historyDecodeTxResult.Result.(*blockchain.TransactionResponse)
		if !ok || historyDecodeTx == nil {
			log.WarnWithContextf(ctx, "解析历史交易结果失败, 跳过: 解码结果类型错误")
			continue
		}

		// 获取交易地址
		exchangeAddress := ""
		for _, vin := range historyDecodeTx.Vin {
			// 检查是否是签名交易且ASM长度合适
			if len(vin.ScriptSig.Asm) > 0 && vin.ScriptSig.Asm[0:2] == "30" && len(vin.ScriptSig.Asm) < 500 {
				// 获取公钥并转换为地址
				if len(vin.ScriptSig.Asm) >= 66 {
					pubkey := vin.ScriptSig.Asm[len(vin.ScriptSig.Asm)-66:]
					addr, err := utility.ConvertCompressedPubkeyToLegacyAddress(pubkey)
					if err == nil {
						exchangeAddress = addr
						break
					}
				}
			}
		}
		historyResponse.ExchangeAddress = exchangeAddress

		// 获取上一个池余额
		var lastFtLpBalance, lastFtABalance, lastTbcBalance int64
		if len(historyDecodeTx.Vin) > 0 &&
			len(historyDecodeTx.Vin[0].ScriptSig.Asm) > 0 &&
			historyDecodeTx.Vin[0].ScriptSig.Asm[0:2] == "30" &&
			len(historyDecodeTx.Vin[0].ScriptSig.Asm) > 500 {

			lastTxid := historyDecodeTx.Vin[0].Txid
			lastTxResult := <-rpcBlockchain.DecodeTxHash(ctx, lastTxid)
			if lastTxResult.Error == nil {
				if lastTx, ok := lastTxResult.Result.(*blockchain.TransactionResponse); ok && lastTx != nil && len(lastTx.Vout) > 1 {
					var err error
					lastFtLpBalance, lastFtABalance, lastTbcBalance, err = utility.GetPoolBalance(lastTx.Vout[1].ScriptPubKey.Asm)
					if err != nil {
						log.WarnWithContextf(ctx, "解析上一个池余额失败: %v", err)
					}
				}
			}
		}

		// 获取当前池余额
		var ftLpBalance, ftABalance, tbcBalance int64
		if len(historyDecodeTx.Vout) > 1 {
			var err error
			ftLpBalance, ftABalance, tbcBalance, err = utility.GetPoolBalance(historyDecodeTx.Vout[1].ScriptPubKey.Asm)
			if err != nil {
				log.WarnWithContextf(ctx, "解析当前池余额失败: %v", err)
				continue
			}
		}

		// 计算余额变化
		ftLpBalanceChange := ftLpBalance - lastFtLpBalance
		ftABalanceChange := ftABalance - lastFtABalance
		tbcBalanceChange := tbcBalance - lastTbcBalance

		historyResponse.FtLpBalanceChange = &ftLpBalanceChange
		historyResponse.TokenPairAPoolBalanceChange = &tbcBalanceChange

		// 获取代币B信息
		if len(historyDecodeTx.Vout) > 1 {
			asmParts := strings.Split(historyDecodeTx.Vout[1].ScriptPubKey.Asm, " ")
			if len(asmParts) > 4 {
				ftContractId := asmParts[4]
				historyResponse.TokenPairBId = ftContractId

				// 获取代币信息
				ftTokensDAO := ft_tokens_dao.NewFtTokensDAO()
				tokenInfo, err := ftTokensDAO.GetFtTokenById(ftContractId)
				if err != nil {
					log.WarnWithContextf(ctx, "获取代币信息失败: %v", err)
				} else if tokenInfo != nil {
					historyResponse.TokenPairBName = tokenInfo.FtName
					historyResponse.TokenPairBDecimal = int(tokenInfo.FtDecimal)
				}

				historyResponse.TokenPairBPoolBalanceChange = &ftABalanceChange
			}
		}

		// 添加到结果列表
		poolHistoryList = append(poolHistoryList, historyResponse)
	}

	log.InfoWithContextf(ctx, "成功获取池历史记录: 池ID=%s, 返回记录数=%d",
		req.PoolId, len(poolHistoryList))

	return poolHistoryList, nil
}
