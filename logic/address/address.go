package address

import (
	"context"
	"fmt"
	"math"
	"math/big"
	"strconv"
	"strings"
	"time"

	"ginproject/entity/blockchain"
	"ginproject/entity/electrumx"
	utility "ginproject/entity/utility"
	"ginproject/middleware/log"
	rpcbchain "ginproject/repo/rpc/blockchain"
	rpcex "ginproject/repo/rpc/electrumx"
)

// AddressLogic 地址业务逻辑
type AddressLogic struct{}

// NewAddressLogic 创建地址业务逻辑实例
func NewAddressLogic() *AddressLogic {
	return &AddressLogic{}
}

// GetAddressUnspentUtxos 获取地址的未花费交易输出
func (l *AddressLogic) GetAddressUnspentUtxos(ctx context.Context, address string) (electrumx.UtxoResponse, error) {
	// 记录开始处理的日志
	log.InfoWithContext(ctx, "开始获取地址的UTXO", "address:", address)

	// 验证地址合法性
	valid, addrType, err := utility.ValidateWIFAddress(address)
	if err != nil || !valid {
		log.ErrorWithContext(ctx, "地址验证失败", "address:", address, "错误:", err)
		return nil, fmt.Errorf("无效的地址格式: %w", err)
	}

	log.InfoWithContext(ctx, "地址验证通过", "address:", address, "type:", addrType)

	// 将地址转换为脚本哈希
	scriptHash, err := utility.AddressToScriptHash(address)
	if err != nil {
		log.ErrorWithContext(ctx, "地址转换为脚本哈希失败", "address:", address, "错误:", err)
		return nil, fmt.Errorf("地址转换失败: %w", err)
	}

	log.InfoWithContext(ctx, "地址已转换为脚本哈希", "address:", address, "scriptHash:", scriptHash)

	// 调用RPC获取UTXO列表
	utxos, err := rpcex.GetListUnspent(scriptHash)
	if err != nil {
		log.ErrorWithContext(ctx, "获取UTXO失败",
			"address:", address,
			"scriptHash:", scriptHash,
			"错误:", err)
		return nil, fmt.Errorf("获取UTXO失败: %w", err)
	}

	log.InfoWithContext(ctx, "成功获取地址UTXO", "address:", address, "count:", len(utxos))
	return utxos, nil
}

// GetAddressHistoryPage 获取地址历史交易信息（支持分页）
func (l *AddressLogic) GetAddressHistoryPage(ctx context.Context, address string, asPage bool, page int) (*electrumx.AddressHistoryResponse, error) {
	// 记录开始处理的日志
	log.InfoWithContext(ctx, "开始获取地址的交易历史(分页模式)",
		"address:", address,
		"asPage:", asPage,
		"page:", page)

	// 验证地址并获取脚本哈希
	scriptHash, err := l.validateAddressAndGetScriptHash(ctx, address)
	if err != nil {
		return nil, err
	}

	// 获取交易历史记录并分页
	historyCount, neededItems, err := l.getPagedHistory(ctx, address, scriptHash, asPage, page)
	if err != nil {
		return nil, err
	}

	// 处理历史记录，创建结果列表
	result := l.processHistoryItems(ctx, address, neededItems)

	// 按时间戳排序
	l.sortHistoryByTimestamp(result)

	// 创建响应
	response := &electrumx.AddressHistoryResponse{
		Address:      address,
		Script:       scriptHash,
		HistoryCount: historyCount,
		Result:       result,
	}

	log.InfoWithContext(ctx, "成功获取地址交易历史(分页模式)",
		"address:", address,
		"asPage:", asPage,
		"page:", page,
		"total_count:", historyCount,
		"returned_count:", len(result))

	return response, nil
}

// validateAddressAndGetScriptHash 验证地址合法性并获取脚本哈希
func (l *AddressLogic) validateAddressAndGetScriptHash(ctx context.Context, address string) (string, error) {
	// 验证地址合法性
	valid, addrType, err := utility.ValidateWIFAddress(address)
	if err != nil || !valid {
		log.ErrorWithContext(ctx, "地址验证失败", "address:", address, "错误:", err)
		return "", fmt.Errorf("无效的地址格式: %w", err)
	}

	log.InfoWithContext(ctx, "地址验证通过", "address:", address, "type:", addrType)

	// 将地址转换为脚本哈希
	scriptHash, err := utility.AddressToScriptHash(address)
	if err != nil {
		log.ErrorWithContext(ctx, "地址转换为脚本哈希失败", "address:", address, "错误:", err)
		return "", fmt.Errorf("地址转换失败: %w", err)
	}

	log.InfoWithContext(ctx, "地址已转换为脚本哈希", "address:", address, "scriptHash:", scriptHash)
	return scriptHash, nil
}

// getPagedHistory 获取历史记录并应用分页
func (l *AddressLogic) getPagedHistory(ctx context.Context, address, scriptHash string, asPage bool, page int) (
	historyCount int,
	neededItems electrumx.ElectrumXHistoryResponse,
	err error,
) {
	// 获取交易历史记录
	historyResponse, err := rpcex.GetScriptHashHistory(scriptHash)
	if err != nil {
		log.ErrorWithContext(ctx, "获取交易历史失败",
			"address:", address,
			"scriptHash:", scriptHash,
			"错误:", err)
		return 0, nil, fmt.Errorf("获取交易历史失败: %w", err)
	}

	// 交易数量
	historyCount = len(historyResponse)
	log.InfoWithContext(ctx, "获取到交易历史记录数",
		"address:", address,
		"count:", historyCount)

	// 反转历史记录顺序（从新到旧）
	for i, j := 0, len(historyResponse)-1; i < j; i, j = i+1, j-1 {
		historyResponse[i], historyResponse[j] = historyResponse[j], historyResponse[i]
	}

	// 根据分页参数获取需要处理的记录
	if asPage {
		start := page * 10
		end := start + 10
		// 确保不会越界
		if start < len(historyResponse) {
			if end > len(historyResponse) {
				end = len(historyResponse)
			}
			neededItems = historyResponse[start:end]
		} else {
			// 超出范围，返回空列表
			log.InfoWithContext(ctx, "请求的页码超出历史记录范围",
				"address:", address,
				"page:", page,
				"total_records:", len(historyResponse))
			neededItems = make(electrumx.ElectrumXHistoryResponse, 0)
		}
	} else {
		if len(historyResponse) > 30 {
			neededItems = historyResponse[:30]
		} else {
			neededItems = historyResponse
		}
	}

	return historyCount, neededItems, nil
}

// processHistoryItems 处理历史交易记录
func (l *AddressLogic) processHistoryItems(ctx context.Context, address string, neededItems electrumx.ElectrumXHistoryResponse) []electrumx.HistoryItem {
	// 创建结果列表
	result := make([]electrumx.HistoryItem, 0, len(neededItems))

	// 处理每条交易记录
	for _, item := range neededItems {
		historyItem, ok := l.processTransactionItem(ctx, address, item)
		if !ok {
			continue
		}
		result = append(result, historyItem)
	}

	return result
}

// processTransactionItem 处理单个交易记录
func (l *AddressLogic) processTransactionItem(ctx context.Context, address string, item electrumx.ElectrumXHistoryItem) (electrumx.HistoryItem, bool) {
	var balanceChange int64
	var totalSpend int64
	var totalReceive int64
	txid := item.TxHash

	// 获取交易详情
	log.InfoWithContext(ctx, "开始获取交易详情", "txid:", txid)
	decodedInfo, err := rpcbchain.DecodeTx(ctx, txid)
	if err != nil {
		log.WarnWithContext(ctx, "获取交易详情失败，跳过此记录",
			"txid:", txid,
			"错误:", err)
		return electrumx.HistoryItem{}, false
	}

	// 创建发送方和接收方集合
	senders := make(map[string]bool)
	receivers := make(map[string]bool)
	ifTypeDetected := false
	txType := "P2PKH"

	// 获取总接收量和接收方
	l.processTransactionOutputs(ctx, address, decodedInfo, &totalReceive, &balanceChange, receivers, &ifTypeDetected, &txType)

	// 获取总支出和发送方
	l.processTransactionInputs(ctx, address, decodedInfo, &totalSpend, &balanceChange, senders)

	// 创建接收方和发送方地址列表以及计算手续费
	recipientAddresses, senderAddresses, feeStr := l.buildAddressesAndFee(address, balanceChange, totalSpend, totalReceive, senders, receivers)

	// 获取时间戳
	utcTime, timeStamp := l.getTransactionTimestamp(ctx, item)

	// 格式化余额变化
	balanceFloatStr := l.formatBalanceChange(balanceChange)

	// 创建历史记录项
	historyItem := electrumx.HistoryItem{
		BalanceChange:      balanceFloatStr,
		BanlanceChange:     balanceFloatStr, // 为了保持与API规范一致
		TxHash:             txid,
		SenderAddresses:    senderAddresses,
		RecipientAddresses: recipientAddresses,
		Fee:                feeStr,
		TimeStamp:          timeStamp,
		UtcTime:            utcTime,
		TxType:             txType,
	}

	return historyItem, true
}

// processTransactionOutputs 处理交易输出
func (l *AddressLogic) processTransactionOutputs(
	ctx context.Context,
	address string,
	decodedInfo *blockchain.TransactionResponse,
	totalReceive *int64,
	balanceChange *int64,
	receivers map[string]bool,
	ifTypeDetected *bool,
	txType *string,
) {
	log.InfoWithContext(ctx, "开始处理交易输出", "address:", address)
	for _, output := range decodedInfo.Vout {
		// 将BTC转换为聪（1 BTC = 1,000,000 聪）
		valueGet := int64(math.Round(output.Value * 1000000))
		*totalReceive += valueGet

		// 处理不同类型的输出脚本
		if output.ScriptPubKey.Type == "pubkeyhash" {
			for _, addr := range output.ScriptPubKey.Addresses {
				receivers[addr] = true
				if addr == address {
					*balanceChange += valueGet
				}
			}
		} else if strings.HasPrefix(output.ScriptPubKey.Asm, "9 OP_PICK OP_TOALTSTACK") {
			if !*ifTypeDetected {
				*ifTypeDetected = true
				*txType = "TBC20"
			}
			if strings.HasSuffix(output.ScriptPubKey.Asm, "01 32436f6465") {
				poolContractID := "Pool_" + output.ScriptPubKey.Asm[len(output.ScriptPubKey.Asm)-53:len(output.ScriptPubKey.Asm)-11]
				receivers[poolContractID] = true
			}
		} else if (strings.HasPrefix(output.ScriptPubKey.Asm, "OP_RETURN") ||
			strings.HasPrefix(output.ScriptPubKey.Asm, "0 OP_RETURN") ||
			strings.HasPrefix(output.ScriptPubKey.Asm, "1 OP_PICK")) && !*ifTypeDetected {
			*ifTypeDetected = true
			*txType = "TBC721"
		} else if strings.HasSuffix(output.ScriptPubKey.Asm, "OP_CHECKMULTISIG") && !*ifTypeDetected {
			*ifTypeDetected = true
			*txType = "P2MS"
			msAddress, err := utility.ConvertP2msScriptToMsAddress(output.ScriptPubKey.Asm)
			if err == nil {
				receivers[msAddress] = true
				if msAddress == address {
					*balanceChange += valueGet
				}
			}
		}
	}
}

// processTransactionInputs 处理交易输入
func (l *AddressLogic) processTransactionInputs(
	ctx context.Context,
	address string,
	decodedInfo *blockchain.TransactionResponse,
	totalSpend *int64,
	balanceChange *int64,
	senders map[string]bool,
) {
	for _, vin := range decodedInfo.Vin {
		if vin.Txid == "" {
			// coinbase交易
			senders["coinbase"] = true
			*totalSpend += 325 // 默认coinbase交易支出
		} else {
			vinTxid := vin.Txid
			vinVout := vin.Vout

			// 获取前一个交易的输出信息
			vinDecoded, err := rpcbchain.DecodeTx(ctx, vinTxid)
			if err != nil {
				log.WarnWithContext(ctx, "获取输入交易详情失败",
					"vin_txid:", vinTxid,
					"错误:", err)
				continue
			}

			if vinVout >= len(vinDecoded.Vout) {
				log.WarnWithContext(ctx, "输入索引超出范围",
					"vin_txid:", vinTxid,
					"vin_vout:", vinVout,
					"vout_length:", len(vinDecoded.Vout))
				continue
			}

			// 将BTC转换为聪
			valueSpend := int64(math.Round(vinDecoded.Vout[vinVout].Value * 1000000))
			*totalSpend += valueSpend

			// 处理不同类型的输入脚本
			if vinDecoded.Vout[vinVout].ScriptPubKey.Type == "pubkeyhash" {
				for _, addr := range vinDecoded.Vout[vinVout].ScriptPubKey.Addresses {
					senders[addr] = true
					if addr == address {
						*balanceChange -= valueSpend
					}
				}
			} else if strings.HasSuffix(vinDecoded.Vout[vinVout].ScriptPubKey.Asm, "OP_CHECKMULTISIG") {
				msAddress, err := utility.ConvertP2msScriptToMsAddress(vinDecoded.Vout[vinVout].ScriptPubKey.Asm)
				if err == nil {
					senders[msAddress] = true
					if msAddress == address {
						*balanceChange -= valueSpend
					}
				}
			} else if strings.HasPrefix(vinDecoded.Vout[vinVout].ScriptPubKey.Asm, "9 OP_PICK OP_TOALTSTACK") {
				if strings.HasSuffix(vinDecoded.Vout[vinVout].ScriptPubKey.Asm, "01 32436f6465") {
					poolContractID := "Pool_" + vinDecoded.Vout[vinVout].ScriptPubKey.Asm[len(vinDecoded.Vout[vinVout].ScriptPubKey.Asm)-53:len(vinDecoded.Vout[vinVout].ScriptPubKey.Asm)-11]
					senders[poolContractID] = true
				}
			}
		}
	}
}

// buildAddressesAndFee 构建地址列表和计算手续费
func (l *AddressLogic) buildAddressesAndFee(
	address string,
	balanceChange int64,
	totalSpend int64,
	totalReceive int64,
	senders map[string]bool,
	receivers map[string]bool,
) (recipientAddresses []string, senderAddresses []string, feeStr string) {
	// 创建接收方和发送方地址列表
	recipientAddresses = make([]string, 0)
	senderAddresses = make([]string, 0)

	// 计算交易手续费
	fee := float64(totalSpend-totalReceive) / 1000000
	feeStr = strconv.FormatFloat(fee, 'f', -1, 64)

	// 确定发送方和接收方
	if balanceChange < 0 {
		senderAddresses = append(senderAddresses, address)
		for receiver := range receivers {
			if receiver != address {
				recipientAddresses = append(recipientAddresses, receiver)
			}
		}
	} else {
		recipientAddresses = append(recipientAddresses, address)
		for sender := range senders {
			if sender != address {
				senderAddresses = append(senderAddresses, sender)
			}
		}
	}

	// 确保发送方和接收方列表不为空
	if len(senderAddresses) == 0 {
		senderAddresses = append(senderAddresses, address)
	}
	if len(recipientAddresses) == 0 {
		recipientAddresses = append(recipientAddresses, address)
	}

	return recipientAddresses, senderAddresses, feeStr
}

// getTransactionTimestamp 获取交易时间戳
func (l *AddressLogic) getTransactionTimestamp(ctx context.Context, item electrumx.ElectrumXHistoryItem) (string, int64) {
	var utcTime string
	var timeStamp int64

	if item.Height < 1 {
		utcTime = "unconfirmed"
	} else {
		blockInfo, err := rpcbchain.GetBlockByHeight(ctx, item.Height)
		if err != nil {
			log.WarnWithContext(ctx, "获取区块信息失败",
				"height:", item.Height,
				"错误:", err)
		} else {
			timeStamp = blockInfo.Time
			utcTime = time.Unix(timeStamp, 0).UTC().Format("2006-01-02 15:04:05")
		}
	}

	return utcTime, timeStamp
}

// formatBalanceChange 格式化余额变化
func (l *AddressLogic) formatBalanceChange(balanceChange int64) string {
	balanceFloat := new(big.Float).Quo(
		new(big.Float).SetInt64(balanceChange),
		new(big.Float).SetInt64(1000000))
	balanceFloatStr := fmt.Sprintf("%+.6f", balanceFloat)
	balanceFloatStr = strings.TrimRight(balanceFloatStr, "0")
	balanceFloatStr = strings.TrimRight(balanceFloatStr, ".")

	if balanceFloatStr == "+" || balanceFloatStr == "" {
		balanceFloatStr = "0"
	}

	return balanceFloatStr
}

// sortHistoryByTimestamp 按时间戳排序历史记录
func (l *AddressLogic) sortHistoryByTimestamp(result []electrumx.HistoryItem) {
	// 按时间戳排序（从新到旧）
	// 未确认的交易（时间戳为0）排在最前面
	for i := 0; i < len(result); i++ {
		for j := i + 1; j < len(result); j++ {
			iValue := result[i].TimeStamp
			jValue := result[j].TimeStamp

			// 如果i是未确认交易，继续
			if iValue == 0 {
				continue
			}

			// 如果j是未确认交易，交换
			if jValue == 0 {
				result[i], result[j] = result[j], result[i]
				continue
			}

			// 按时间戳降序排序
			if iValue < jValue {
				result[i], result[j] = result[j], result[i]
			}
		}
	}
}

// GetAddressBalance 获取地址余额
func (l *AddressLogic) GetAddressBalance(ctx context.Context, address string) (*electrumx.AddressBalanceResponse, error) {
	// 记录开始处理的日志
	log.InfoWithContext(ctx, "开始获取地址余额", "address:", address)

	// 验证地址并获取脚本哈希
	scriptHash, err := l.validateAddressAndGetScriptHash(ctx, address)
	if err != nil {
		log.ErrorWithContext(ctx, "获取地址余额失败：地址验证错误", "address:", address, "错误:", err)
		return nil, err
	}

	// 调用RPC获取余额
	balanceResponse, err := rpcex.GetBalance(scriptHash)
	if err != nil {
		log.ErrorWithContext(ctx, "获取地址余额失败：RPC调用错误",
			"address:", address,
			"scriptHash:", scriptHash,
			"错误:", err)
		return nil, fmt.Errorf("获取地址余额失败: %w", err)
	}

	// 计算总余额
	totalBalance := balanceResponse.Confirmed + balanceResponse.Unconfirmed

	// 创建响应
	response := &electrumx.AddressBalanceResponse{
		Balance:     totalBalance,
		Confirmed:   balanceResponse.Confirmed,
		Unconfirmed: balanceResponse.Unconfirmed,
	}

	log.InfoWithContext(ctx, "成功获取地址余额",
		"address:", address,
		"confirmed:", balanceResponse.Confirmed,
		"unconfirmed:", balanceResponse.Unconfirmed,
		"total:", totalBalance)

	return response, nil
}

// GetAddressFrozenBalance 获取地址冻结余额
func (l *AddressLogic) GetAddressFrozenBalance(ctx context.Context, address string) (*electrumx.FrozenBalanceResponse, error) {
	// 记录开始处理的日志
	log.InfoWithContext(ctx, "开始获取地址冻结余额", "address:", address)

	// 验证地址并获取脚本哈希
	scriptHash, err := l.validateAddressAndGetScriptHash(ctx, address)
	if err != nil {
		log.ErrorWithContext(ctx, "获取地址冻结余额失败：地址验证错误", "address:", address, "错误:", err)
		return nil, err
	}

	// 调用RPC获取冻结余额
	frozenBalanceResponse, err := rpcex.GetAddressFrozenBalance(ctx, address)
	if err != nil {
		log.ErrorWithContext(ctx, "获取地址冻结余额失败：RPC调用错误",
			"address:", address,
			"scriptHash:", scriptHash,
			"错误:", err)
		return nil, fmt.Errorf("获取地址冻结余额失败: %w", err)
	}

	log.InfoWithContext(ctx, "成功获取地址冻结余额",
		"address:", address,
		"frozen:", frozenBalanceResponse.Frozen)

	return frozenBalanceResponse, nil
}
