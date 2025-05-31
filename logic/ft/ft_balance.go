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

// GetFtBalance 获取FT余额
func (l *FtLogic) GetFtBalance(ctx context.Context, req *ft.FtBalanceAddressRequest) (*ft.FtBalanceAddressResponse, error) {
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

	// 调用DAO层获取余额
	ftBalance, err := l.ftTxoDAO.GetTotalBalanceByHolder(ctx, combineScript, req.ContractId)
	if err != nil {
		log.ErrorWithContextf(ctx, "查询FT余额失败: %v", err)
		return nil, fmt.Errorf("查询FT余额失败: %v", err)
	}

	log.InfoWithContextf(ctx, "FT余额查询成功: %d", ftBalance)

	// 构造响应
	response := &ft.FtBalanceAddressResponse{
		CombineScript: combineScript,
		FtContractId:  req.ContractId,
		FtDecimal:     int(ftDecimal),
		FtBalance:     ftBalance,
	}

	return response, nil
}

// GetMultiFtBalanceByAddress 获取地址持有的多个代币余额
func (l *FtLogic) GetMultiFtBalanceByAddress(ctx context.Context, req *ft.FtBalanceMultiContractRequest) ([]ft.TBC20FTBalanceResponse, error) {
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

	// 初始化响应结果切片
	responseList := make([]ft.TBC20FTBalanceResponse, 0, len(req.FtContractId))

	// 遍历每个合约ID，查询余额
	for _, contractId := range req.FtContractId {
		// 获取代币小数位数
		ftDecimal, err := l.ftTokensDAO.GetFtDecimalByContractId(ctx, contractId)
		if err != nil {
			log.ErrorWithContextf(ctx, "获取代币小数位数失败，合约ID=%s: %v", contractId, err)
			// 跳过错误的合约，继续处理其他合约
			continue
		}

		// 调用DAO层获取余额
		ftBalance, err := l.ftTxoDAO.GetTotalBalanceByHolder(ctx, combineScript, contractId)
		if err != nil {
			log.ErrorWithContextf(ctx, "查询FT余额失败，合约ID=%s: %v", contractId, err)
			// 跳过错误的合约，继续处理其他合约
			continue
		}

		// 构造单个代币的余额响应
		response := ft.TBC20FTBalanceResponse{
			CombineScript: combineScript,
			FtContractId:  contractId,
			FtDecimal:     int(ftDecimal),
			FtBalance:     ftBalance,
		}

		// 添加到响应列表
		responseList = append(responseList, response)
	}

	log.InfoWithContextf(ctx, "批量查询FT余额成功，合约数量: %d", len(responseList))
	return responseList, nil
}

// getTokenListByCombineScript 获取代币列表的通用逻辑
func (l *FtLogic) getTokenListByCombineScript(ctx context.Context, combineScript string) ([]ft.TokenInfo, error) {
	

	log.InfoWithContextf(ctx, "通过合并脚本获取代币列表请求: 合并脚本=%s", combineScript)

	// 获取该地址持有的所有代币合约ID
	contractIds, err := l.ftTxoDAO.GetFtContractIdsByHolder(ctx, combineScript)
	if err != nil {
		log.ErrorWithContextf(ctx, "获取地址持有的代币合约ID列表失败: %v", err)
		return nil, fmt.Errorf("获取地址持有的代币合约ID列表失败: %v", err)
	}

	if len(contractIds) == 0 {
		log.InfoWithContextf(ctx, "地址[%s]未持有任何代币", combineScript)
		return []ft.TokenInfo{}, nil
	}

	log.InfoWithContextf(ctx, "地址[%s]持有的代币数量: %d", combineScript, len(contractIds))

	// 初始化代币列表
	tokenList := make([]ft.TokenInfo, 0, len(contractIds))

	// 遍历所有合约ID，查询详细信息
	for _, contractId := range contractIds {
		// 查询代币详细信息
		token, err := l.ftTokensDAO.GetFtTokenById(contractId)
		if err != nil {
			log.WarnWithContextf(ctx, "获取代币[%s]详细信息失败: %v，跳过此代币", contractId, err)
			continue
		}

		// 跳过LP代币（流动性代币）
		if token.FtSymbol == "LP" || token.FtName == "LP Token" {
			log.InfoWithContextf(ctx, "跳过LP代币: %s", contractId)
			continue
		}

		// 获取代币余额
		ftBalance, err := l.ftTxoDAO.GetTotalBalanceByHolder(ctx, combineScript, contractId)
		if err != nil {
			log.WarnWithContextf(ctx, "获取代币[%s]余额失败: %v，跳过此代币", contractId, err)
			continue
		}

		// 如果余额为0，则跳过
		if ftBalance == 0 {
			log.InfoWithContextf(ctx, "跳过余额为0的代币: %s", contractId)
			continue
		}

		// 添加代币信息到列表
		tokenInfo := ft.TokenInfo{
			FtContractId: contractId,
			FtDecimal:    int(token.FtDecimal),
			FtBalance:    ftBalance,
			FtName:       token.FtName,
			FtSymbol:     token.FtSymbol,
		}
		tokenList = append(tokenList, tokenInfo)
	}

	return tokenList, nil
}

func (l *FtLogic) GetTokensListHeldByCombineScript(ctx context.Context, req *ft.TBC20TokenListHeldByCombineScriptRequest) (*ft.TBC20TokenListHeldByCombineScriptResponse, error) {
	// 使用entity层的验证逻辑
	if err := req.Validate(); err != nil {
		log.ErrorWithContextf(ctx, "参数验证失败: %v", err)
		return nil, fmt.Errorf("参数验证失败: %v", err)
	}

	tokenList, err := l.getTokenListByCombineScript(ctx, req.CombineScript)
	if err != nil {
		return nil, err
	}

	// 构造响应
	response := &ft.TBC20TokenListHeldByCombineScriptResponse{
		CombineScript: req.CombineScript,
		TokenCount:    len(tokenList),
		TokenList:     tokenList,
	}

	log.InfoWithContextf(ctx, "成功获取地址[%s]持有的代币列表，数量: %d", req.CombineScript, len(tokenList))
	return response, nil
}

func (l *FtLogic) GetTokensListHeldByAddress(ctx context.Context, req *ft.TBC20TokenListHeldByAddressRequest) (*ft.TBC20TokenListHeldByAddressResponse, error) {
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

	tokenList, err := l.getTokenListByCombineScript(ctx, combineScript)
	if err != nil {
		return nil, err
	}

	// 构造响应
	response := &ft.TBC20TokenListHeldByAddressResponse{
		Address:    req.Address,
		TokenCount: len(tokenList),
		TokenList:  tokenList,
	}

	log.InfoWithContextf(ctx, "成功获取地址[%s]持有的代币列表，数量: %d", req.Address, len(tokenList))
	return response, nil
}

// GetFtBalanceByCombineScript 根据合并脚本和合约哈希获取FT余额
func (l *FtLogic) GetFtBalanceByCombineScript(ctx context.Context, req *ft.FtBalanceCombineScriptRequest) (*ft.FtBalanceCombineScriptResponse, error) {
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

	log.InfoWithContextf(ctx, "根据合并脚本获取FT余额: 合并脚本=%s, 合约哈希=%s", combineScript, req.ContractHash)
	// 获取代币小数位数
	ftDecimal, err := l.ftTokensDAO.GetFtDecimalByContractId(ctx, req.ContractHash)
	if err != nil {
		log.ErrorWithContextf(ctx, "获取代币小数位数失败: %v", err)
		return nil, fmt.Errorf("获取代币小数位数失败: %v", err)
	}
	// 从数据库获取FT余额
	dbBalance, err := l.getFtBalanceFromDB(ctx, combineScript, req.ContractHash)
	if err == nil {
		log.InfoWithContextf(ctx, "从数据库成功获取FT余额: %d", dbBalance)
		// 构造响应
		return &ft.FtBalanceCombineScriptResponse{
			CombineScript: combineScript,
			ContractHash:  req.ContractHash,
			FtDecimal:     int(ftDecimal),
			FtBalance:     dbBalance,
		}, nil
	}

	log.WarnWithContextf(ctx, "从数据库获取FT余额失败: %v, 尝试通过RPC获取", err)

	// 方法2: 通过RPC获取数据（备用方案）
	// 1. 解码合约交易，获取合约脚本
	decodeContractResult := <-repoBlockchain.DecodeTxHash(ctx, req.ContractHash)
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

	// 获取合约脚本
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

	// 4. 计算总余额
	var contractBalance uint64 = 0

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

		// 获取脚本
		tapeScript := utxoTx.Vout[vout].ScriptPubKey.Hex

		// 提取FT值部分（根据示例为脚本的第6至102个字符）
		if len(tapeScript) >= 102 {
			valueHex := tapeScript[6:102]

			// 分段处理8字节的FT值
			var outputAmount uint64 = 0
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
				outputAmount += segmentValue
			}

			contractBalance += outputAmount
		}
	}

	log.InfoWithContextf(ctx, "通过RPC获取FT余额成功: %d", contractBalance)

	// 构造响应
	response := &ft.FtBalanceCombineScriptResponse{
		CombineScript: combineScript,
		ContractHash:  req.ContractHash,
		FtDecimal:     int(ftDecimal),
		FtBalance:     contractBalance,
	}

	return response, nil
}

// getFtBalanceFromDB 从数据库获取FT余额
func (l *FtLogic) getFtBalanceFromDB(ctx context.Context, script string, contractHash string) (uint64, error) {
	log.InfoWithContextf(ctx, "从数据库获取FT余额: 脚本=%s, 合约哈希=%s", script, contractHash)

	// 1. 首先需要将合约哈希转换为合约ID (contractHash通常是交易ID)
	contractId := contractHash

	// 2. 查询此脚本和合约ID的总余额
	balance, err := l.ftTxoDAO.GetTotalBalanceByHolder(ctx, script, contractId)
	if err != nil {
		log.ErrorWithContextf(ctx, "从数据库查询FT余额失败: %v", err)
		return 0, fmt.Errorf("查询FT余额失败: %v", err)
	}

	log.InfoWithContextf(ctx, "从数据库获取FT余额成功: %d", balance)
	return balance, nil
}
