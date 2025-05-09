package ft

import (
	"context"
	"fmt"

	entityElectrumx "ginproject/entity/electrumx"
	"ginproject/entity/ft"
	"ginproject/entity/utility"
	"ginproject/middleware/log"
	repoBlockchain "ginproject/repo/rpc/blockchain"
	repoElectrumx "ginproject/repo/rpc/electrumx"
)

// GetTokenHistory 获取代币历史交易记录
func (l *FtLogic) GetTokenHistory(ctx context.Context, req *ft.FtTokenHistoryRequest) (*ft.FtTokenHistoryResponse, error) {
	// 参数验证
	if err := ft.ValidateFtTokenHistoryRequest(req); err != nil {
		log.ErrorWithContextf(ctx, "参数验证失败: %v", err)
		return nil, fmt.Errorf("参数验证失败: %v", err)
	}

	// 获取代币代码脚本
	ftCodeScript, err := l.ftTokensDAO.GetFtCodeScript(ctx, req.FtContractId)
	if err != nil {
		log.ErrorWithContextf(ctx, "获取代币代码脚本失败: %v", err)
		return nil, fmt.Errorf("获取代币代码脚本失败: %v", err)
	}

	// 检查代码脚本是否为空
	if ftCodeScript == "" {
		log.WarnWithContextf(ctx, "未找到代币信息: contractId=%s", req.FtContractId)
		return &ft.FtTokenHistoryResponse{}, nil
	}

	// 将代码脚本转换为脚本哈希
	ftCodeScriptHash, err := utility.ConvertStrToSha256(ftCodeScript)
	if err != nil {
		log.ErrorWithContextf(ctx, "转换脚本哈希失败: %v", err)
		return nil, fmt.Errorf("转换脚本哈希失败: %v", err)
	}

	log.InfoWithContextf(ctx, "成功获取代币代码脚本哈希: contractId=%s, hash=%s",
		req.FtContractId, ftCodeScriptHash)

	// 获取脚本历史交易列表
	ftHistoryTxs, err := repoElectrumx.GetScriptHashHistory(ctx, ftCodeScriptHash)
	if err != nil {
		log.ErrorWithContextf(ctx, "获取脚本历史失败: %v", err)
		return nil, fmt.Errorf("获取脚本历史失败: %v", err)
	}

	// 反转历史记录顺序（从最新到最旧）
	totalTxs := len(ftHistoryTxs)
	log.InfoWithContextf(ctx, "获取脚本历史成功，共 %d 条记录", totalTxs)

	if totalTxs == 0 {
		return &ft.FtTokenHistoryResponse{}, nil
	}

	// 分页处理
	start := req.Page * req.Size
	end := (req.Page + 1) * req.Size
	if start >= totalTxs {
		log.InfoWithContextf(ctx, "请求的页码超出范围: page=%d, size=%d, total=%d",
			req.Page, req.Size, totalTxs)
		return &ft.FtTokenHistoryResponse{}, nil
	}

	if end > totalTxs {
		end = totalTxs
	}

	// 对所有交易记录进行反转
	reversedTxs := make([]entityElectrumx.ElectrumXHistoryItem, totalTxs)
	for i, tx := range ftHistoryTxs {
		reversedTxs[totalTxs-1-i] = tx
	}

	// 获取分页后的交易记录
	pageTxs := reversedTxs[start:end]
	log.InfoWithContextf(ctx, "分页处理成功: page=%d, size=%d, 当前页记录数=%d",
		req.Page, req.Size, len(pageTxs))

	// 解析每个交易历史记录
	historyList := make([]ft.FtTokenHistoryItem, 0, len(pageTxs))
	for _, historyItem := range pageTxs {
		// 解析交易
		txInfo, err := l.decodeTxHistory(ctx, historyItem.TxHash)
		if err != nil {
			log.WarnWithContextf(ctx, "解析交易历史失败: txid=%s, error=%v", historyItem.TxHash, err)
			continue
		}

		// 添加到历史列表
		historyList = append(historyList, ft.FtTokenHistoryItem{
			Txid:         historyItem.TxHash,
			FtContractId: req.FtContractId,
			TxInfo:       txInfo,
		})
	}

	// 构建响应
	response := ft.FtTokenHistoryResponse(historyList)

	log.InfoWithContextf(ctx, "成功获取代币历史交易记录: contractId=%s, total=%d, 返回记录数=%d",
		req.FtContractId, totalTxs, len(historyList))

	return &response, nil
}

// decodeTxHistory 解析交易历史
func (l *FtLogic) decodeTxHistory(ctx context.Context, txid string) (*ft.FtTxDecodeResponse, error) {
	// 调用区块链RPC获取交易详情
	decodeTx, err := repoBlockchain.DecodeTx(ctx, txid)
	if err != nil {
		log.ErrorWithContextf(ctx, "获取交易详情失败: %v", err)
		return nil, fmt.Errorf("获取交易详情失败: %v", err)
	}

	// 初始化输入输出列表
	inputList := make([]ft.FtTxDecodeData, 0, len(decodeTx.Vin))
	outputList := make([]ft.FtTxDecodeData, 0, len(decodeTx.Vout))

	// 处理输入
	for vinIndex, vin := range decodeTx.Vin {
		// 获取输入UTXO的FT信息
		ftBalance, ftHolderScript, ftContractId, err := l.ftTxoDAO.GetFtUtxoInfo(
			ctx, vin.Txid, vin.Vout)
		if err != nil || ftContractId == "" {
			// 非FT交易或查询失败，跳过
			continue
		}

		// 处理地址
		address := ""
		if len(ftHolderScript) >= 2 && ftHolderScript[len(ftHolderScript)-2:] == "00" {
			// 普通地址
			address, err = utility.ConvertCombineScriptToAddress(ftHolderScript)
			if err != nil {
				log.WarnWithContextf(ctx, "转换地址失败: %v", err)
				continue
			}
		} else if len(ftHolderScript) >= 2 && ftHolderScript[len(ftHolderScript)-2:] == "01" {
			// 多签名或池控制
			if vinIndex > 0 && len(decodeTx.Vin) > vinIndex-1 &&
				decodeTx.Vin[vinIndex-1].ScriptSig.Asm != "" &&
				len(decodeTx.Vin[vinIndex-1].ScriptSig.Asm) > 2 &&
				decodeTx.Vin[vinIndex-1].ScriptSig.Asm[:2] == "0 " {
				// 多签
				address, err = utility.ConvertP2msUnlockScriptToAddress(decodeTx.Vin[vinIndex-1].ScriptSig.Asm)
				if err != nil {
					log.WarnWithContextf(ctx, "转换多签地址失败: %v", err)
					address = "Pool_" + ftHolderScript
				}
			} else {
				// 池控制
				address = "Pool_" + ftHolderScript
			}
		}

		// 获取代币小数位数
		ftDecimal, err := l.ftTokensDAO.GetFtDecimalByContractId(ctx, ftContractId)
		if err != nil {
			log.WarnWithContextf(ctx, "获取代币小数位数失败: %v", err)
			ftDecimal = 0
		}

		// 添加到输入列表
		inputList = append(inputList, ft.FtTxDecodeData{
			Txid:       vin.Txid,
			Vout:       vin.Vout,
			Address:    address,
			ContractId: ftContractId,
			FtBalance:  int64(ftBalance),
			FtDecimal:  int(ftDecimal),
		})
	}

	// 处理输出
	for _, vout := range decodeTx.Vout {
		// 获取输出UTXO的FT信息
		ftBalance, ftHolderScript, ftContractId, err := l.ftTxoDAO.GetFtUtxoInfo(
			ctx, txid, vout.N)
		if err != nil || ftContractId == "" {
			// 非FT交易或查询失败，跳过
			continue
		}

		// 处理地址
		address := ""
		if len(ftHolderScript) >= 2 && ftHolderScript[len(ftHolderScript)-2:] == "00" {
			// 普通地址
			address, err = utility.ConvertCombineScriptToAddress(ftHolderScript)
			if err != nil {
				log.WarnWithContextf(ctx, "转换地址失败: %v", err)
				continue
			}
		} else if len(ftHolderScript) >= 2 && ftHolderScript[len(ftHolderScript)-2:] == "01" {
			// 多签名或池控制
			// 尝试匹配输入的池地址
			foundPool := false
			for _, input := range inputList {
				if input.Address != "" && len(input.Address) > 5 &&
					input.Address[:5] == "Pool_" &&
					input.Address == "Pool_"+ftHolderScript {
					address = "Pool_" + ftHolderScript
					foundPool = true
					break
				}
			}
			if !foundPool {
				address = "Pool_or_MS_" + ftHolderScript
			}
		}

		// 获取代币小数位数
		ftDecimal, err := l.ftTokensDAO.GetFtDecimalByContractId(ctx, ftContractId)
		if err != nil {
			log.WarnWithContextf(ctx, "获取代币小数位数失败: %v", err)
			ftDecimal = 0
		}

		// 添加到输出列表
		outputList = append(outputList, ft.FtTxDecodeData{
			Txid:       txid,
			Vout:       vout.N,
			Address:    address,
			ContractId: ftContractId,
			FtBalance:  int64(ftBalance),
			FtDecimal:  int(ftDecimal),
		})
	}

	// 构建交易解码响应
	txDecodeResponse := &ft.FtTxDecodeResponse{
		Txid:   txid,
		Input:  inputList,
		Output: outputList,
	}

	return txDecodeResponse, nil
}
