package ft

import (
	"context"
	"fmt"
	"strconv"

	"ginproject/entity/blockchain"
	"ginproject/entity/ft"
	"ginproject/entity/utility"
	"ginproject/middleware/log"
	repoBlockchain "ginproject/repo/rpc/blockchain"
	"ginproject/repo/rpc/electrumx"
)

// GetFtUtxosByAddress 获取FT UTXO列表
func (l *FtLogic) GetFtUtxosByAddress(ctx context.Context, req *ft.FtUtxoAddressRequest) (*ft.FtUtxoAddressResponse, error) {
	// 使用entity层的验证逻辑
	if err := req.Validate(); err != nil {
		log.ErrorWithContextf(ctx, "参数验证失败: %v", err)
		return nil, fmt.Errorf("参数验证失败: %v", err)
	}

	// 从请求对象中获取组合脚本
	combineScript, err := req.GetCombineScript()
	if err != nil {
		log.ErrorWithContextf(ctx, "获取组合脚本失败: %v", err)
		return nil, fmt.Errorf("获取组合脚本失败: %v", err)
	}

	// 添加00作为校验
	combineScript += "00"

	// 获取代币小数位数
	ftDecimal, err := l.ftTokensDAO.GetFtDecimalByContractId(ctx, req.ContractId)
	if err != nil {
		log.ErrorWithContextf(ctx, "获取代币小数位数失败: %v", err)
		return nil, fmt.Errorf("获取代币小数位数失败: %v", err)
	}

	// 调用DAO层获取未花费的UTXO列表
	utxos, err := l.ftTxoDAO.GetUnspentFtTxosByHolderAndContract(combineScript, req.ContractId)
	if err != nil {
		log.ErrorWithContextf(ctx, "查询FT UTXO列表失败: %v", err)
		return nil, fmt.Errorf("查询FT UTXO列表失败: %v", err)
	}

	// 构造响应
	response := &ft.FtUtxoAddressResponse{
		FtUtxoList: make([]*ft.FtUtxoItem, 0, len(utxos)),
	}

	// 遍历UTXO列表，构造UTXO项
	for _, utxo := range utxos {
		utxoItem := &ft.FtUtxoItem{
			UtxoId:       utxo.UtxoTxid,
			UtxoVout:     utxo.UtxoVout,
			UtxoBalance:  utxo.UtxoBalance,
			FtContractId: utxo.FtContractId,
			FtDecimal:    int(ftDecimal),
			FtBalance:    utxo.FtBalance,
		}
		response.FtUtxoList = append(response.FtUtxoList, utxoItem)
	}

	log.InfoWithContextf(ctx, "FT UTXO查询成功: 共%d条记录", len(response.FtUtxoList))

	return response, nil
}

// GetFtUtxosByCombineScript 根据合并脚本和合约ID获取FT UTXO列表
func (l *FtLogic) GetFtUtxosByCombineScript(ctx context.Context, req *ft.FtUtxoCombineScriptRequest) (*ft.TBC20FTUtxoResponse, error) {
	// 使用entity层的验证逻辑
	if err := req.Validate(); err != nil {
		log.ErrorWithContextf(ctx, "参数验证失败: %v", err)
		return nil, fmt.Errorf("参数验证失败: %v", err)
	}

	// 添加00作为校验（如果需要）
	combineScript := req.CombineScript
	if len(combineScript) > 0 && combineScript[len(combineScript)-2:] != "00" {
		combineScript += "00"
	}

	log.InfoWithContextf(ctx, "根据合并脚本获取FT UTXO: 合并脚本=%s, 合约ID=%s", combineScript, req.ContractId)

	
	dbResponse, err := l.getFtUtxoFromDB(ctx, combineScript, req.ContractId)
	if err == nil && len(dbResponse.Data) > 0 {
		log.InfoWithContextf(ctx, "从数据库成功获取FT UTXO: 共%d条记录", len(dbResponse.Data))
		return dbResponse, nil
	}

	if err != nil {
		log.WarnWithContextf(ctx, "从数据库获取FT UTXO失败: %v, 尝试通过RPC获取", err)
	} else {
		log.WarnWithContextf(ctx, "数据库中未找到任何记录，尝试通过RPC获取")
	}

	// 方法2: 通过RPC获取数据（备用方案）
	// 1. 解码合约交易，获取合约脚本
	decodeContractResult := <-repoBlockchain.DecodeTxHash(ctx, req.ContractId)
	if decodeContractResult.Error != nil {
		log.ErrorWithContextf(ctx, "解码合约交易失败: %v", decodeContractResult.Error)
		return nil, fmt.Errorf("解码合约交易失败: %v", decodeContractResult.Error)
	}

	contractTx, ok := decodeContractResult.Result.(*blockchain.TransactionResponse)
	if !ok {
		return nil, fmt.Errorf("解码合约交易响应格式错误")
	}

	if len(contractTx.Vout) == 0 {
		return nil, fmt.Errorf("合约交易输出为空")
	}

	// 获取合约脚本特征
	codeScriptHex := contractTx.Vout[0].ScriptPubKey.Hex
	contractTrait := codeScriptHex[0 : len(codeScriptHex)-54]
	completeScript := contractTrait + combineScript + "0502436f6465" // "0502436f6465"是"Code"的十六进制表示

	// 2. 计算脚本哈希
	scriptHash, err := utility.ConvertStrToSha256(completeScript)
	if err != nil {
		log.ErrorWithContextf(ctx, "计算脚本哈希失败: %v", err)
		return nil, fmt.Errorf("计算脚本哈希失败: %v", err)
	}

	// 3. 获取未花费的UTXO列表
	unspentUtxos, err := electrumx.GetUnspent(ctx, scriptHash)
	if err != nil {
		log.ErrorWithContextf(ctx, "获取未花费UTXO失败: %v", err)
		return nil, fmt.Errorf("获取未花费UTXO失败: %v", err)
	}

	// 4. 构建响应
	response := &ft.TBC20FTUtxoResponse{
		Data: make([]*ft.TBC20FTUtxoItem, 0, len(unspentUtxos)),
	}

	for _, utxo := range unspentUtxos {
		txid := utxo.TxHash
		vout := int(utxo.TxPos) + 1 // ElectrumX索引从0开始，需要加1

		// 解码未花费的交易
		decodeUtxoTxResult := <-repoBlockchain.DecodeTxHash(ctx, txid)
		if decodeUtxoTxResult.Error != nil {
			log.WarnWithContextf(ctx, "解码UTXO交易失败，跳过: %v", decodeUtxoTxResult.Error)
			continue
		}

		utxoTx, ok := decodeUtxoTxResult.Result.(*blockchain.TransactionResponse)
		if !ok || len(utxoTx.Vout) <= vout {
			log.WarnWithContextf(ctx, "UTXO交易输出格式错误或索引超出范围，跳过")
			continue
		}

		// 获取脚本和Satoshi值（比特币值）
		tapeScript := utxoTx.Vout[vout].ScriptPubKey.Hex
		satoshiValue := uint64(utxoTx.Vout[vout].Value * 100000000) // 浮点数转换为比特币中的Satoshi值

		// 提取FT值部分（根据示例为脚本的第6至102个字符）
		var ftAmount uint64 = 0
		if len(tapeScript) >= 102 {
			valueHex := tapeScript[6:102]

			// 分段处理8字节的FT值
			for i := 0; i < len(valueHex); i += 16 {
				if i+16 > len(valueHex) {
					break
				}
				segment := valueHex[i : i+16]

				// 字节序反转（每两个字符为一个字节）
				var reversedSegment string
				for j := 14; j >= 0; j -= 2 {
					reversedSegment += segment[j : j+2]
				}

				// 转换为数值并累加
				segmentValue, err := strconv.ParseUint(reversedSegment, 16, 64)
				if err != nil {
					log.WarnWithContextf(ctx, "解析FT值段失败: %v", err)
					continue
				}
				ftAmount += segmentValue
			}
		}

		// 获取代币小数位数
		ftDecimal, err := l.ftTokensDAO.GetFtDecimalByContractId(ctx, req.ContractId)
		if err != nil {
			log.ErrorWithContextf(ctx, "获取代币小数位数失败: %v", err)
			return nil, fmt.Errorf("获取代币小数位数失败: %v", err)
		}

		// 构建UTXO项
		utxoItem := &ft.TBC20FTUtxoItem{
			UtxoId:       txid,
			UtxoVout:     int(utxo.TxPos),
			UtxoBalance:  satoshiValue,
			FtContractId: req.ContractId,
			FtDecimal:    int(ftDecimal),
			FtBalance:    ftAmount,
		}

		response.Data = append(response.Data, utxoItem)
	}

	log.InfoWithContextf(ctx, "通过RPC成功获取FT UTXO: 共%d条记录", len(response.Data))
	return response, nil
}

// getFtUtxoFromDB 从数据库获取FT UTXO列表
func (l *FtLogic) getFtUtxoFromDB(ctx context.Context, combineScript string, contractId string) (*ft.TBC20FTUtxoResponse, error) {
	log.InfoWithContextf(ctx, "从数据库获取FT UTXO数据: 合并脚本=%s, 合约ID=%s", combineScript, contractId)

	// 获取代币小数位数 (虽然这里不直接使用，但响应中可能需要)
	_, err := l.ftTokensDAO.GetFtDecimalByContractId(ctx, contractId)
	if err != nil {
		log.ErrorWithContextf(ctx, "获取代币小数位数失败: %v", err)
		return nil, fmt.Errorf("获取代币小数位数失败: %v", err)
	}

	// 获取未花费的UTXO列表
	utxos, err := l.ftTxoDAO.GetUnspentFtTxosByHolderAndContract(combineScript, contractId)
	if err != nil {
		log.ErrorWithContextf(ctx, "从数据库查询FT UTXO列表失败: %v", err)
		return nil, fmt.Errorf("查询FT UTXO列表失败: %v", err)
	}

	// 构造响应
	response := &ft.TBC20FTUtxoResponse{
		Data: make([]*ft.TBC20FTUtxoItem, 0, len(utxos)),
	}

	// 将DAO返回的数据转换为API响应格式
	for _, utxo := range utxos {
		// 获取代币小数位数
		ftDecimal, err := l.ftTokensDAO.GetFtDecimalByContractId(ctx, utxo.FtContractId)
		if err != nil {
			log.ErrorWithContextf(ctx, "获取代币小数位数失败: %v", err)
			return nil, fmt.Errorf("获取代币小数位数失败: %v", err)
		}
		utxoItem := &ft.TBC20FTUtxoItem{
			UtxoId:       utxo.UtxoTxid,
			UtxoVout:     utxo.UtxoVout,
			UtxoBalance:  utxo.UtxoBalance,
			FtContractId: utxo.FtContractId,
			FtDecimal:    int(ftDecimal),
			FtBalance:    utxo.FtBalance,
		}
		response.Data = append(response.Data, utxoItem)
	}

	log.InfoWithContextf(ctx, "从数据库获取FT UTXO成功: 共%d条记录", len(response.Data))
	return response, nil
}
