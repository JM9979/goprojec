package address

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"

	"ginproject/entity/dbtable"
	"ginproject/entity/electrumx"
	utility "ginproject/entity/utility"
	"ginproject/middleware/log"
	"ginproject/repo/db/address_transactions_dao"
	"ginproject/repo/db/transaction_participants_dao"
	"ginproject/repo/db/transactions_dao"
	rpcex "ginproject/repo/rpc/electrumx"
)

// AsyncUtxoResult 异步UTXO结果
type AsyncUtxoResult struct {
	Utxos electrumx.UtxoResponse
	Error error
}

// AddressLogic 地址业务逻辑
type AddressLogic struct{}

// NewAddressLogic 创建地址业务逻辑实例
func NewAddressLogic() *AddressLogic {
	return &AddressLogic{}
}

// GetAddressUnspentUtxos 获取地址的未花费交易输出（异步版本）
func (l *AddressLogic) GetAddressUnspentUtxos(ctx context.Context, address string) chan *AsyncUtxoResult {
	resultChan := make(chan *AsyncUtxoResult, 1)

	go func() {
		defer close(resultChan)

		// 记录开始处理的日志
		log.InfoWithContext(ctx, "开始获取地址的UTXO", "address:", address)

		// 验证地址合法性
		valid, addrType, err := utility.ValidateWIFAddress(address)
		if err != nil || !valid {
			log.ErrorWithContext(ctx, "地址验证失败", "address:", address, "错误:", err)
			resultChan <- &AsyncUtxoResult{
				Error: fmt.Errorf("无效的地址格式: %w", err),
			}
			return
		}

		log.InfoWithContext(ctx, "地址验证通过", "address:", address, "type:", addrType)

		// 将地址转换为脚本哈希
		scriptHash, err := utility.AddressToScriptHash(address)
		if err != nil {
			log.ErrorWithContext(ctx, "地址转换为脚本哈希失败", "address:", address, "错误:", err)
			resultChan <- &AsyncUtxoResult{
				Error: fmt.Errorf("地址转换失败: %w", err),
			}
			return
		}

		log.InfoWithContext(ctx, "地址已转换为脚本哈希", "address:", address, "scriptHash:", scriptHash)

		// 调用RPC获取UTXO列表
		utxos, err := rpcex.GetListUnspent(ctx, scriptHash)
		if err != nil {
			log.ErrorWithContext(ctx, "获取UTXO失败",
				"address:", address,
				"scriptHash:", scriptHash,
				"错误:", err)
			resultChan <- &AsyncUtxoResult{
				Error: fmt.Errorf("获取UTXO失败: %w", err),
			}
			return
		}

		log.InfoWithContext(ctx, "成功获取地址UTXO", "address:", address, "count:", len(utxos))
		resultChan <- &AsyncUtxoResult{
			Utxos: utxos,
		}
	}()

	return resultChan
}

// GetAddressHistoryPage 获取地址历史交易信息（支持分页）
func (l *AddressLogic) GetAddressHistoryPage(ctx context.Context, address string, asPage bool, page int) (*electrumx.AddressHistoryResponse, error) {
	// 记录开始处理的日志
	log.InfoWithContext(ctx, "开始获取地址的交易历史(数据库异步模式)",
		"address:", address,
		"asPage:", asPage,
		"page:", page)

	// 创建上下文，允许取消操作
	ctxWithCancel, cancel := context.WithCancel(ctx)
	defer cancel()

	// 验证地址并获取脚本哈希
	scriptHash, err := l.validateAddressAndGetScriptHash(ctx, address)
	if err != nil {
		return nil, err
	}

	// 设置分页参数
	limit := 10
	if !asPage {
		limit = 30
	}
	offset := page * limit

	// 创建错误通道，用于中断所有正在进行的查询
	errChan := make(chan error, 3)

	// 创建等待组，等待所有查询完成
	var wg sync.WaitGroup
	wg.Add(1)

	// 统计该地址的交易总数
	var historyCount int64
	var countErr error

	go func() {
		defer wg.Done()
		historyCount, countErr = address_transactions_dao.CountAddressTransactions(ctxWithCancel, address)
		if countErr != nil {
			errChan <- countErr
		}
	}()

	// 等待统计结果
	wg.Wait()

	// 检查错误
	if countErr != nil {
		log.ErrorWithContext(ctx, "统计地址交易数量失败",
			"address:", address,
			"错误:", countErr)
		return nil, fmt.Errorf("统计地址交易数量失败: %w", countErr)
	}

	// 查询该地址的交易列表（分页）
	addrTxs, err := address_transactions_dao.GetAddressTransactions(ctxWithCancel, address, offset, limit)
	if err != nil {
		log.ErrorWithContext(ctx, "查询地址交易记录失败",
			"address:", address,
			"offset:", offset,
			"limit:", limit,
			"错误:", err)
		return nil, fmt.Errorf("查询地址交易记录失败: %w", err)
	}

	// 如果没有交易记录，返回空结果
	if len(addrTxs) == 0 {
		log.InfoWithContext(ctx, "地址没有交易记录",
			"address:", address,
			"offset:", offset,
			"limit:", limit)
		return &electrumx.AddressHistoryResponse{
			Address:      address,
			Script:       scriptHash,
			HistoryCount: int(historyCount),
			Result:       []electrumx.HistoryItem{},
		}, nil
	}

	// 提取交易哈希列表
	txHashes := make([]string, 0, len(addrTxs))
	for _, tx := range addrTxs {
		txHashes = append(txHashes, tx.TxHash)
	}

	// 创建等待组，等待所有查询完成
	wg = sync.WaitGroup{}
	wg.Add(2)

	// 定义存储结果的变量
	var txDetails []*dbtable.Transaction
	var participants []*dbtable.TransactionParticipant
	var txDetailsErr, participantsErr error

	// 查询交易详情
	go func() {
		defer wg.Done()
		txDetails, txDetailsErr = transactions_dao.GetTransactionsByTxHashes(ctxWithCancel, txHashes)
		if txDetailsErr != nil {
			errChan <- txDetailsErr
		}
	}()

	// 查询交易参与方
	go func() {
		defer wg.Done()
		participants, participantsErr = transaction_participants_dao.GetParticipantsByTxHashes(ctxWithCancel, txHashes)
		if participantsErr != nil {
			errChan <- participantsErr
		}
	}()

	// 等待所有查询完成
	wg.Wait()

	// 检查查询错误
	if txDetailsErr != nil {
		log.ErrorWithContext(ctx, "查询交易详情失败",
			"address:", address,
			"txHashes:", txHashes,
			"错误:", txDetailsErr)
		return nil, fmt.Errorf("查询交易详情失败: %w", txDetailsErr)
	}

	if participantsErr != nil {
		log.ErrorWithContext(ctx, "查询交易参与方失败",
			"address:", address,
			"txHashes:", txHashes,
			"错误:", participantsErr)
		return nil, fmt.Errorf("查询交易参与方失败: %w", participantsErr)
	}

	// 整理数据，构建响应
	result := l.buildHistoryItemsFromDB(ctx, address, addrTxs, txDetails, participants)

	// 按时间戳排序
	l.sortHistoryByTimestamp(result)

	// 创建响应
	response := &electrumx.AddressHistoryResponse{
		Address:      address,
		Script:       scriptHash,
		HistoryCount: int(historyCount),
		Result:       result,
	}

	log.InfoWithContext(ctx, "成功获取地址交易历史(数据库异步模式)",
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

// sortHistoryByTimestamp 按时间戳排序历史记录
func (l *AddressLogic) sortHistoryByTimestamp(result []electrumx.HistoryItem) {
	// 使用sort.Slice实现排序，与Python版本保持一致的排序逻辑
	sort.Slice(result, func(i, j int) bool {
		iValue := result[i].TimeStamp
		jValue := result[j].TimeStamp

		// 如果i是未确认交易而j不是，i排前面
		if iValue == 0 && jValue != 0 {
			return true
		}

		// 如果j是未确认交易而i不是，j排前面
		if jValue == 0 && iValue != 0 {
			return false
		}

		// 都是确认交易或都是未确认交易，按时间戳降序
		return iValue > jValue
	})
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
	balanceResponse, err := rpcex.GetBalance(ctx, scriptHash)
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

// buildHistoryItemsFromDB 从数据库查询结果构建历史记录项
func (l *AddressLogic) buildHistoryItemsFromDB(
	ctx context.Context,
	address string,
	addrTxs []*dbtable.AddressTransaction,
	txDetails []*dbtable.Transaction,
	participants []*dbtable.TransactionParticipant,
) []electrumx.HistoryItem {
	// 将交易详情转换为map，方便查询
	txMap := make(map[string]*dbtable.Transaction)
	for _, tx := range txDetails {
		txMap[tx.TxHash] = tx
	}

	// 将交易参与方转换为map，方便查询
	participantMap := make(map[string]map[dbtable.Role][]string)
	for _, p := range participants {
		if _, ok := participantMap[p.TxHash]; !ok {
			participantMap[p.TxHash] = make(map[dbtable.Role][]string)
		}
		participantMap[p.TxHash][p.Role] = append(participantMap[p.TxHash][p.Role], p.Address)
	}

	// 构建历史记录项
	result := make([]electrumx.HistoryItem, 0, len(addrTxs))
	for _, addrTx := range addrTxs {
		// 查找交易详情
		tx, ok := txMap[addrTx.TxHash]
		if !ok {
			log.WarnWithContext(ctx, "未找到交易详情，跳过此记录",
				"address:", address,
				"txHash:", addrTx.TxHash)
			continue
		}

		// 获取发送方和接收方地址列表
		senderAddresses := []string{}
		if senders, ok := participantMap[addrTx.TxHash][dbtable.RoleSender]; ok {
			senderAddresses = senders
		}

		recipientAddresses := []string{}
		if recipients, ok := participantMap[addrTx.TxHash][dbtable.RoleRecipient]; ok {
			recipientAddresses = recipients
		}

		// 确保发送方和接收方列表不为空
		if len(senderAddresses) == 0 {
			senderAddresses = append(senderAddresses, address)
		}
		if len(recipientAddresses) == 0 {
			recipientAddresses = append(recipientAddresses, address)
		}

		// 格式化余额变化
		balanceChange := fmt.Sprintf("%+.6f", addrTx.BalanceChange)
		balanceChange = strings.TrimRight(balanceChange, "0")
		balanceChange = strings.TrimRight(balanceChange, ".")
		if balanceChange == "+" || balanceChange == "" {
			balanceChange = "0"
		}

		// 格式化手续费
		feeStr := fmt.Sprintf("%.8f", tx.Fee)
		feeStr = strings.TrimRight(feeStr, "0")
		feeStr = strings.TrimRight(feeStr, ".")

		// 创建历史记录项
		historyItem := electrumx.HistoryItem{
			BalanceChange:      balanceChange,
			TxHash:             addrTx.TxHash,
			SenderAddresses:    senderAddresses,
			RecipientAddresses: recipientAddresses,
			Fee:                feeStr,
			TimeStamp:          tx.TimeStamp,
			UtcTime:            tx.UtcTime,
			TxType:             tx.TxType,
		}

		result = append(result, historyItem)
	}

	return result
}
