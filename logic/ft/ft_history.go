package ft

import (
	"context"
	"fmt"
	"time"

	"ginproject/entity/ft"
	"ginproject/entity/utility"
	"ginproject/middleware/log"
	rpcblockchain "ginproject/repo/rpc/blockchain"
	"ginproject/repo/rpc/electrumx"
)

// GetFtHistory 获取地址的代币交易历史
// 实现了基于ElectrumX的交易历史查询，并从FT UTXO集合中提取相关的代币余额变化信息
// 支持分页查询，返回符合需求格式的交易历史
func (l *FtLogic) GetFtHistory(ctx context.Context, req *ft.FtHistoryRequest) (*ft.FtHistoryResponse, error) {
	// 参数校验
	if err := req.Validate(); err != nil {
		log.ErrorWithContextf(ctx, "参数验证失败: %v", err)
		return nil, fmt.Errorf("参数验证失败: %v", err)
	}

	// 获取FT代币脚本信息
	ftLockingScript, err := l.getFtLockingScript(ctx, req.ContractId, req.Address)
	if err != nil {
		return nil, err
	}

	// 获取交易历史
	historyResult, err := l.fetchTransactionHistory(ctx, ftLockingScript)
	if err != nil {
		return nil, err
	}

	// 对历史记录进行分页处理
	totalCount := len(historyResult)
	pagedHistory := l.paginateHistory(historyResult, req.Page, req.Size)

	// 处理历史记录详情
	historyList, err := l.processFtHistoryDetails(ctx, pagedHistory, req.ContractId, req.Address)
	if err != nil {
		return nil, err
	}

	// 构造最终响应
	response := &ft.FtHistoryResponse{
		Address:      req.Address,
		ScriptHash:   ftLockingScript,
		HistoryCount: totalCount,
		Result:       historyList,
	}

	log.InfoWithContextf(ctx, "获取FT交易历史成功，合约ID=%s，地址=%s，历史数量=%d",
		req.ContractId, req.Address, len(historyList))

	return response, nil
}

// getFtLockingScript 获取FT代币锁定脚本
func (l *FtLogic) getFtLockingScript(ctx context.Context, contractId, address string) (string, error) {
	log.InfoWithContextf(ctx, "开始获取FT代币锁定脚本，合约ID=%s，地址=%s", contractId, address)

	// 获取代币代码脚本和精度
	ftCodeScript, _, err := l.ftTokensDAO.GetFtCodeScriptAndDecimal(ctx, contractId)
	if err != nil {
		log.ErrorWithContextf(ctx, "获取代币信息失败，合约ID=%s: %v", contractId, err)
		return "", fmt.Errorf("获取代币信息失败: %v", err)
	}

	if ftCodeScript == "" {
		log.ErrorWithContextf(ctx, "代币不存在，合约ID=%s", contractId)
		return "", fmt.Errorf("代币不存在，合约ID=%s", contractId)
	}

	// 获取组合脚本
	pubKeyHash, err := utility.ConvertAddressToPublicKeyHash(address)
	if err != nil {
		log.ErrorWithContextf(ctx, "地址转换为公钥哈希失败: %v", err)
		return "", fmt.Errorf("地址转换为公钥哈希失败: %v", err)
	}

	// 添加00作为校验
	combineScript := pubKeyHash + "00"
	log.DebugWithContextf(ctx, "生成组合脚本: %s", combineScript)

	// 构造FT代码脚本
	// 从代码脚本中截取合适的部分，然后插入组合脚本
	if len(ftCodeScript) < 54 {
		log.ErrorWithContextf(ctx, "代币代码脚本格式不正确: %s", ftCodeScript)
		return "", fmt.Errorf("代币代码脚本格式不正确")
	}
	ftCodeScript = ftCodeScript[:len(ftCodeScript)-54] + combineScript + ftCodeScript[len(ftCodeScript)-12:]
	log.DebugWithContextf(ctx, "构造FT代码脚本: %s", ftCodeScript)

	// 计算锁定脚本哈希
	ftLockingScript, err := utility.ConvertStrToSha256(ftCodeScript)
	if err != nil {
		log.ErrorWithContextf(ctx, "计算锁定脚本哈希失败: %v", err)
		return "", fmt.Errorf("计算锁定脚本哈希失败: %v", err)
	}

	log.InfoWithContextf(ctx, "成功获取FT代币锁定脚本: %s", ftLockingScript)
	return ftLockingScript, nil
}

// fetchTransactionHistory 获取交易历史记录
func (l *FtLogic) fetchTransactionHistory(ctx context.Context, scriptHash string) ([]interface{}, error) {
	log.InfoWithContextf(ctx, "开始从ElectrumX获取脚本哈希历史: %s", scriptHash)
	historyResult, err := electrumx.GetTransactionHistory(scriptHash)
	if err != nil {
		log.ErrorWithContextf(ctx, "获取交易历史失败: %v", err)
		return nil, fmt.Errorf("获取交易历史失败: %v", err)
	}

	log.InfoWithContextf(ctx, "成功获取交易历史记录，条数=%d", len(historyResult))

	// 反转历史记录顺序（按时间降序排列）
	for i, j := 0, len(historyResult)-1; i < j; i, j = i+1, j-1 {
		historyResult[i], historyResult[j] = historyResult[j], historyResult[i]
	}

	return historyResult, nil
}

// paginateHistory 对历史记录进行分页处理
func (l *FtLogic) paginateHistory(history []interface{}, page, size int) []interface{} {
	start := page * size
	end := start + size

	if start >= len(history) {
		return []interface{}{}
	} else if end > len(history) {
		return history[start:]
	}

	return history[start:end]
}

// processFtHistoryDetails 处理FT历史记录详情
func (l *FtLogic) processFtHistoryDetails(ctx context.Context, historyItems []interface{},
	contractId, address string) ([]ft.FtHistoryRecord, error) {

	log.InfoWithContextf(ctx, "开始处理FT历史记录详情，合约ID=%s，地址=%s，条数=%d",
		contractId, address, len(historyItems))

	// 获取代币精度
	_, ftDecimal, err := l.ftTokensDAO.GetFtCodeScriptAndDecimal(ctx, contractId)
	if err != nil {
		log.ErrorWithContextf(ctx, "获取代币精度失败，合约ID=%s: %v", contractId, err)
		return nil, fmt.Errorf("获取代币精度失败: %v", err)
	}
	log.DebugWithContextf(ctx, "代币精度: %d", ftDecimal)

	// 获取组合脚本
	pubKeyHash, err := utility.ConvertAddressToPublicKeyHash(address)
	if err != nil {
		log.ErrorWithContextf(ctx, "地址转换为公钥哈希失败: %v", err)
		return nil, fmt.Errorf("地址转换为公钥哈希失败: %v", err)
	}
	combineScript := pubKeyHash + "00"
	log.DebugWithContextf(ctx, "生成组合脚本: %s", combineScript)

	historyList := make([]ft.FtHistoryRecord, 0, len(historyItems))

	// 处理每条历史记录
	for i, item := range historyItems {
		log.DebugWithContextf(ctx, "处理历史记录项 #%d", i+1)
		record, err := l.processHistoryItem(ctx, item, contractId, address, combineScript, int32(ftDecimal))
		if err != nil {
			log.WarnWithContextf(ctx, "处理历史记录#%d失败: %v", i+1, err)
			continue
		}

		if record != nil {
			historyList = append(historyList, *record)
			log.DebugWithContextf(ctx, "历史记录#%d处理成功, 交易ID=%s", i+1, record.TxId)
		}
	}

	log.InfoWithContextf(ctx, "FT历史记录详情处理完成，成功处理%d/%d条记录",
		len(historyList), len(historyItems))
	return historyList, nil
}

// processHistoryItem 处理单个历史记录项
func (l *FtLogic) processHistoryItem(ctx context.Context, item interface{}, contractId, address,
	combineScript string, ftDecimal int32) (*ft.FtHistoryRecord, error) {

	// 将item转换为map类型
	historyItem, ok := item.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("历史记录格式不正确")
	}

	// 提取交易哈希和区块高度
	txHash, ok := historyItem["tx_hash"].(string)
	if !ok {
		return nil, fmt.Errorf("交易哈希格式不正确")
	}

	height, ok := historyItem["height"].(float64)
	if !ok {
		return nil, fmt.Errorf("区块高度格式不正确")
	}

	log.InfoWithContextf(ctx, "处理交易历史: txHash=%s, height=%d", txHash, int(height))

	// 获取时间戳和UTC时间
	timeStamp, utcTime := l.getTxTimeInfo(ctx, height)
	log.DebugWithContextf(ctx, "交易时间信息: timeStamp=%d, utcTime=%s", timeStamp, utcTime)

	// 获取交易详情
	decodeTx, err := rpcblockchain.CallRPC("getrawtransaction", []interface{}{txHash, 1}, false)
	if err != nil {
		log.ErrorWithContextf(ctx, "获取交易详情失败: txHash=%s, 错误=%v", txHash, err)
		return nil, fmt.Errorf("获取交易详情失败: %v", err)
	}

	decodeTxMap, ok := decodeTx.(map[string]interface{})
	if !ok {
		log.ErrorWithContextf(ctx, "交易详情格式不正确: %v", decodeTx)
		return nil, fmt.Errorf("交易详情格式不正确")
	}

	// 处理交易输入输出
	log.DebugWithContextf(ctx, "开始处理交易详情: txHash=%s", txHash)
	txInfo, err := l.processTxDetails(ctx, decodeTxMap, txHash, contractId, combineScript)
	if err != nil {
		log.ErrorWithContextf(ctx, "处理交易详情失败: txHash=%s, 错误=%v", txHash, err)
		return nil, err
	}

	// 构建发送方和接收方地址列表
	senderList, recipientList := l.buildAddressLists(address, txInfo.ftBalanceChange,
		txInfo.senderAddresses, txInfo.recipientAddresses)

	log.DebugWithContextf(ctx, "交易地址信息: 发送方=%v, 接收方=%v", senderList, recipientList)
	log.DebugWithContextf(ctx, "FT余额变化: %d, 交易费用: %f",
		txInfo.ftBalanceChange, float64(txInfo.txTotalSpend-txInfo.txTotalReceive)/1000000)

	// 添加历史记录
	return &ft.FtHistoryRecord{
		TxId:                   txHash,
		FtContractId:           contractId,
		FtBalanceChange:        txInfo.ftBalanceChange,
		FtDecimal:              int(ftDecimal),
		TxFee:                  float64(txInfo.txTotalSpend-txInfo.txTotalReceive) / 1000000,
		SenderCombineScript:    senderList,
		RecipientCombineScript: recipientList,
		TimeStamp:              timeStamp,
		UtcTime:                utcTime,
	}, nil
}

// getTxTimeInfo 获取交易的时间信息
func (l *FtLogic) getTxTimeInfo(ctx context.Context, height float64) (int64, string) {
	if height < 1 {
		// 未确认的交易
		return 0, "unconfirmed"
	}

	// 已确认的交易，获取区块信息
	blockInfo, err := rpcblockchain.CallRPC("getblockbyheight", []interface{}{int(height)}, false)
	if err != nil {
		log.WarnWithContextf(ctx, "获取区块信息失败: %v", err)
		return 0, ""
	}

	blockInfoMap, ok := blockInfo.(map[string]interface{})
	if !ok {
		return 0, ""
	}

	// 提取时间戳
	if ts, ok := blockInfoMap["time"].(float64); ok {
		timeStamp := int64(ts)
		utcTime := time.Unix(timeStamp, 0).UTC().Format("2006-01-02 15:04:05")
		return timeStamp, utcTime
	}

	return 0, ""
}

// 交易信息结构
type transactionInfo struct {
	txTotalSpend       uint64
	txTotalReceive     uint64
	ftBalanceChange    int64
	senderAddresses    map[string]struct{}
	recipientAddresses map[string]struct{}
}

// processTxDetails 处理交易详情
func (l *FtLogic) processTxDetails(ctx context.Context, decodeTxMap map[string]interface{},
	txHash, contractId, combineScript string) (*transactionInfo, error) {

	log.InfoWithContextf(ctx, "开始处理交易输入输出: txHash=%s", txHash)

	txInfo := &transactionInfo{
		senderAddresses:    make(map[string]struct{}),
		recipientAddresses: make(map[string]struct{}),
	}

	// 处理交易输入
	err := l.processTxInputs(ctx, decodeTxMap, txInfo, contractId, combineScript)
	if err != nil {
		log.ErrorWithContextf(ctx, "处理交易输入失败: %v", err)
		return nil, err
	}

	// 处理交易输出
	err = l.processTxOutputs(ctx, decodeTxMap, txHash, txInfo, contractId, combineScript)
	if err != nil {
		log.ErrorWithContextf(ctx, "处理交易输出失败: %v", err)
		return nil, err
	}

	// 移除1BitcoinEaterAddressDontSendf59kuE地址
	delete(txInfo.recipientAddresses, "1BitcoinEaterAddressDontSendf59kuE")

	log.DebugWithContextf(ctx, "交易处理完成: 总花费=%d, 总接收=%d, FT余额变化=%d",
		txInfo.txTotalSpend, txInfo.txTotalReceive, txInfo.ftBalanceChange)
	return txInfo, nil
}

// processTxInputs 处理交易输入
func (l *FtLogic) processTxInputs(ctx context.Context, decodeTxMap map[string]interface{},
	txInfo *transactionInfo, contractId, combineScript string) error {

	vinArray, ok := decodeTxMap["vin"].([]interface{})
	if !ok {
		return fmt.Errorf("交易输入格式不正确")
	}

	log.DebugWithContextf(ctx, "开始处理%d个交易输入", len(vinArray))

	for vinIndex, vin := range vinArray {
		vinMap, ok := vin.(map[string]interface{})
		if !ok {
			continue
		}

		// 处理Coinbase交易
		if _, hasCoinbase := vinMap["coinbase"]; hasCoinbase {
			txInfo.txTotalSpend += 325 // 固定coinbase值
			log.DebugWithContextf(ctx, "处理Coinbase交易输入")
			continue
		}

		// 获取输入交易ID和输出索引
		vinTxid, ok := vinMap["txid"].(string)
		if !ok {
			continue
		}

		vinVout, ok := vinMap["vout"].(float64)
		if !ok {
			continue
		}

		log.DebugWithContextf(ctx, "处理交易输入[%d]: txid=%s, vout=%d",
			vinIndex, vinTxid, int(vinVout))

		// 获取输入交易详情
		vinDecodeTx, err := rpcblockchain.CallRPC("getrawtransaction", []interface{}{vinTxid, 1}, false)
		if err != nil {
			log.WarnWithContextf(ctx, "获取输入交易详情失败: %v", err)
			continue
		}

		vinDecodeTxMap, ok := vinDecodeTx.(map[string]interface{})
		if !ok {
			continue
		}

		// 处理输入的值和FT信息
		l.processInputValue(ctx, vinDecodeTxMap, int(vinVout), vinTxid, vinMap, txInfo,
			contractId, combineScript, vinIndex)
	}

	return nil
}

// processInputValue 处理输入的值和FT信息
func (l *FtLogic) processInputValue(ctx context.Context, vinDecodeTxMap map[string]interface{},
	vinVout int, vinTxid string, vinMap map[string]interface{}, txInfo *transactionInfo,
	contractId, combineScript string, vinIndex int) {

	// 获取输入交易的输出
	voutArray, ok := vinDecodeTxMap["vout"].([]interface{})
	if !ok || vinVout >= len(voutArray) {
		log.WarnWithContextf(ctx, "输入交易输出格式不正确或索引越界: txid=%s, vout=%d", vinTxid, vinVout)
		return
	}

	voutMap, ok := voutArray[vinVout].(map[string]interface{})
	if !ok {
		return
	}

	// 获取输入金额
	value, ok := voutMap["value"].(float64)
	if !ok {
		return
	}

	txInfo.txTotalSpend += uint64(value * 1000000)
	log.DebugWithContextf(ctx, "输入金额: %f BSV", value)

	// 获取FT UTXO信息
	ftBalance, ftHolderScript, ftContractId, err := l.ftTxoDAO.GetFtUtxoInfo(ctx, vinTxid, vinVout)
	if err != nil {
		log.WarnWithContextf(ctx, "获取FT UTXO信息失败: txid=%s, vout=%d, 错误=%v",
			vinTxid, vinVout, err)
		return
	}

	if ftHolderScript == "" || ftContractId == "" {
		log.DebugWithContextf(ctx, "非FT相关UTXO: txid=%s, vout=%d", vinTxid, vinVout)
		return
	}

	log.DebugWithContextf(ctx, "发现FT UTXO: 余额=%d, 合约ID=%s, 脚本=%s",
		ftBalance, ftContractId, ftHolderScript)

	// 处理发送方地址
	if ftHolderScript == combineScript && ftContractId == contractId {
		// 自己的地址，记录余额变化
		txInfo.ftBalanceChange -= int64(ftBalance)
		log.DebugWithContextf(ctx, "检测到FT支出: -%d", ftBalance)
	}

	l.processAddressFromScript(ctx, ftHolderScript, vinMap, txInfo.senderAddresses, vinIndex)
}

// processTxOutputs 处理交易输出
func (l *FtLogic) processTxOutputs(ctx context.Context, decodeTxMap map[string]interface{},
	txHash string, txInfo *transactionInfo, contractId, combineScript string) error {

	voutArray, ok := decodeTxMap["vout"].([]interface{})
	if !ok {
		return fmt.Errorf("交易输出格式不正确")
	}

	log.DebugWithContextf(ctx, "开始处理%d个交易输出", len(voutArray))

	for voutIndex, vout := range voutArray {
		voutMap, ok := vout.(map[string]interface{})
		if !ok {
			continue
		}

		// 获取输出金额
		value, ok := voutMap["value"].(float64)
		if !ok {
			continue
		}

		txInfo.txTotalReceive += uint64(value * 1000000)

		// 获取输出索引
		n, ok := voutMap["n"].(float64)
		if !ok {
			continue
		}

		log.DebugWithContextf(ctx, "处理交易输出[%d]: txid=%s, n=%d, 金额=%f BSV",
			voutIndex, txHash, int(n), value)

		// 获取FT UTXO信息
		ftBalance, ftHolderScript, ftContractId, err := l.ftTxoDAO.GetFtUtxoInfo(ctx, txHash, int(n))
		if err != nil {
			log.WarnWithContextf(ctx, "获取FT UTXO信息失败: txid=%s, n=%d, 错误=%v",
				txHash, int(n), err)
			continue
		}

		if ftHolderScript == "" || ftContractId == "" {
			log.DebugWithContextf(ctx, "非FT相关UTXO: txid=%s, n=%d", txHash, int(n))
			continue
		}

		log.DebugWithContextf(ctx, "发现FT UTXO: 余额=%d, 合约ID=%s, 脚本=%s",
			ftBalance, ftContractId, ftHolderScript)

		// 处理接收方地址
		if ftHolderScript == combineScript && ftContractId == contractId {
			// 自己的地址，记录余额变化
			txInfo.ftBalanceChange += int64(ftBalance)
			log.DebugWithContextf(ctx, "检测到FT收入: +%d", ftBalance)
		}

		l.processOutputAddress(ftHolderScript, txInfo)
	}

	return nil
}

// processAddressFromScript 从脚本中提取地址信息
func (l *FtLogic) processAddressFromScript(ctx context.Context, ftHolderScript string,
	vinMap map[string]interface{}, addressMap map[string]struct{}, vinIndex int) {

	if ftHolderScript == "" {
		return
	}

	log.DebugWithContextf(ctx, "处理脚本地址: %s", ftHolderScript)

	if ftHolderScript[len(ftHolderScript)-2:] == "00" {
		// 普通地址
		address, err := utility.ConvertCombineScriptToAddress(ftHolderScript)
		if err != nil {
			log.WarnWithContextf(ctx, "转换组合脚本为地址失败: %v", err)
			return
		}

		addressMap[address] = struct{}{}
		log.DebugWithContextf(ctx, "识别普通地址: %s", address)
	} else if ftHolderScript[len(ftHolderScript)-2:] == "01" {
		// 多签地址或池控制地址
		scriptSig, ok := vinMap["scriptSig"].(map[string]interface{})
		if !ok {
			log.WarnWithContextf(ctx, "脚本签名格式不正确")
			return
		}

		asm, ok := scriptSig["asm"].(string)
		if !ok {
			log.WarnWithContextf(ctx, "ASM格式不正确")
			return
		}

		// 判断是否为多签地址
		if vinIndex > 0 && asm[:2] == "0 " {
			// 多签地址
			address, err := utility.ConvertP2msUnlockScriptToAddress(asm)
			if err != nil {
				log.WarnWithContextf(ctx, "转换P2MS解锁脚本为地址失败: %v", err)
				return
			}

			addressMap[address] = struct{}{}
			log.DebugWithContextf(ctx, "识别多签地址: %s", address)
		} else {
			// 池控制地址
			poolAddress := "Pool_" + ftHolderScript
			addressMap[poolAddress] = struct{}{}
			log.DebugWithContextf(ctx, "识别池控制地址: %s", poolAddress)
		}
	}
}

// processOutputAddress 处理输出地址
func (l *FtLogic) processOutputAddress(ftHolderScript string, txInfo *transactionInfo) {
	if ftHolderScript == "" {
		return
	}

	log.DebugWithContextf(context.Background(), "处理输出脚本地址: %s", ftHolderScript)

	if ftHolderScript[len(ftHolderScript)-2:] == "00" {
		// 普通地址
		address, err := utility.ConvertCombineScriptToAddress(ftHolderScript)
		if err != nil {
			log.WarnWithContextf(context.Background(), "转换组合脚本为地址失败: %v", err)
			return
		}

		txInfo.recipientAddresses[address] = struct{}{}
		log.DebugWithContextf(context.Background(), "识别接收方普通地址: %s", address)
	} else if ftHolderScript[len(ftHolderScript)-2:] == "01" {
		// 池控制或多签地址
		if _, ok := txInfo.senderAddresses["Pool_"+ftHolderScript]; ok {
			// 已知的池控制地址
			poolAddress := "Pool_" + ftHolderScript
			txInfo.recipientAddresses[poolAddress] = struct{}{}
			log.DebugWithContextf(context.Background(), "识别接收方池控制地址: %s", poolAddress)
		} else {
			// 未知的池控制或多签地址
			msAddress := "Pool_or_MS_" + ftHolderScript
			txInfo.recipientAddresses[msAddress] = struct{}{}
			log.DebugWithContextf(context.Background(), "识别接收方池控制或多签地址: %s", msAddress)
		}
	}
}

// buildAddressLists 构建发送方和接收方地址列表
func (l *FtLogic) buildAddressLists(queryAddress string, ftBalanceChange int64,
	senderAddresses, recipientAddresses map[string]struct{}) ([]string, []string) {

	senderList := make([]string, 0, len(senderAddresses))
	recipientList := make([]string, 0, len(recipientAddresses))

	if ftBalanceChange < 0 {
		// 发送交易
		senderList = append(senderList, queryAddress)
		for addr := range recipientAddresses {
			if addr != queryAddress {
				recipientList = append(recipientList, addr)
			}
		}
	} else if ftBalanceChange > 0 {
		// 接收交易
		recipientList = append(recipientList, queryAddress)
		for addr := range senderAddresses {
			if addr != queryAddress {
				senderList = append(senderList, addr)
			}
		}
	} else {
		// 自己给自己转账或其他情况
		senderList = append(senderList, queryAddress)
		recipientList = append(recipientList, queryAddress)
	}

	return senderList, recipientList
}
