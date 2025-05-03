package ft

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"ginproject/entity/ft"
	"ginproject/entity/utility"
	"ginproject/middleware/log"
	"ginproject/repo/rpc/blockchain"
	"ginproject/repo/rpc/electrumx"
)

// GetPoolHistoryByPoolId 根据池子ID获取历史记录
func (l *FtLogic) GetPoolHistoryByPoolId(ctx context.Context, req *ft.TBC20PoolHistoryRequest) ([]ft.TBC20PoolHistoryResponse, error) {
	log.InfoWithContextf(ctx, "获取池子历史记录逻辑处理: 池子ID=%s, 页码=%d, 每页大小=%d", req.PoolId, req.Page, req.Size)

	// 参数验证
	if err := req.Validate(); err != nil {
		log.ErrorWithContextf(ctx, "请求参数验证失败: %v", err)
		return nil, err
	}

	// 使用新实现的详细方法
	return l.GetPoolHistoryPageSize(ctx, req.PoolId, req.Page, req.Size)
}

// GetPoolHistoryPageSize 根据池子ID获取历史记录的详细实现
// 此方法通过区块链RPC和ElectrumX查询池子的交易历史，并进行详细的分析处理
func (l *FtLogic) GetPoolHistoryPageSize(ctx context.Context, poolId string, page int, size int) ([]ft.TBC20PoolHistoryResponse, error) {
	log.InfoWithContextf(ctx, "开始获取池子历史详细信息: 池子ID=%s, 页码=%d, 每页大小=%d", poolId, page, size)

	// 1. 获取池子交易详情
	txMap, err := l.getPoolTransaction(ctx, poolId)
	if err != nil {
		return nil, err
	}
	if txMap == nil {
		return []ft.TBC20PoolHistoryResponse{}, nil
	}

	// 2. 获取池子脚本哈希
	scriptHash, err := l.getPoolScriptHash(ctx, txMap)
	if err != nil {
		return nil, err
	}
	if scriptHash == "" {
		return []ft.TBC20PoolHistoryResponse{}, nil
	}

	// 3. 获取脚本历史记录
	pagedHistory, err := l.getPagedScriptHistory(ctx, scriptHash, page, size)
	if err != nil {
		return nil, err
	}

	// 4. 处理每条历史记录
	var poolHistoryListResult []ft.TBC20PoolHistoryResponse
	for _, tx := range pagedHistory {
		// 获取历史交易详情
		historyItem, err := l.processHistoryTransaction(ctx, tx["tx_hash"].(string), poolId)
		if err != nil {
			log.WarnWithContextf(ctx, "处理历史交易失败: %v", err)
			continue
		}

		poolHistoryListResult = append(poolHistoryListResult, historyItem)
	}

	log.InfoWithContextf(ctx, "获取池子历史记录成功: 池子ID=%s, 历史记录数=%d", poolId, len(poolHistoryListResult))
	return poolHistoryListResult, nil
}

// getPoolTransaction 获取池子交易详情
func (l *FtLogic) getPoolTransaction(ctx context.Context, poolId string) (map[string]interface{}, error) {
	decodeTx, err := blockchain.DecodeRawTransaction(ctx, poolId)
	if err != nil {
		log.ErrorWithContextf(ctx, "获取池子交易详情失败: %v", err)
		return nil, err
	}

	// 确保交易数据有效
	if decodeTx == nil {
		log.ErrorWithContextf(ctx, "获取到的池子交易详情为空")
		return nil, nil
	}

	// 将interface{}转换为map以便访问
	txMap, ok := decodeTx.(map[string]interface{})
	if !ok {
		log.ErrorWithContextf(ctx, "交易数据类型转换失败")
		return nil, fmt.Errorf("交易数据类型转换失败")
	}

	return txMap, nil
}

// getPoolScriptHash 获取池子脚本哈希
func (l *FtLogic) getPoolScriptHash(ctx context.Context, txMap map[string]interface{}) (string, error) {
	voutArray, ok := txMap["vout"].([]interface{})
	if !ok || len(voutArray) == 0 {
		log.ErrorWithContextf(ctx, "交易输出数据无效")
		return "", nil
	}

	// 获取第一个输出的脚本
	firstVout, ok := voutArray[0].(map[string]interface{})
	if !ok {
		log.ErrorWithContextf(ctx, "无法解析第一个交易输出")
		return "", nil
	}

	scriptPubKey, ok := firstVout["scriptPubKey"].(map[string]interface{})
	if !ok {
		log.ErrorWithContextf(ctx, "无法解析scriptPubKey")
		return "", nil
	}

	hex, ok := scriptPubKey["hex"].(string)
	if !ok {
		log.ErrorWithContextf(ctx, "无法获取脚本十六进制")
		return "", nil
	}

	// 计算脚本哈希
	scriptHash, err := utility.ConvertStrToSha256(hex)
	if err != nil {
		log.ErrorWithContextf(ctx, "计算脚本哈希失败: %v", err)
		return "", err
	}

	return scriptHash, nil
}

// getPagedScriptHistory 获取分页的脚本历史记录
func (l *FtLogic) getPagedScriptHistory(ctx context.Context, scriptHash string, page int, size int) ([]map[string]interface{}, error) {
	scriptHistory, err := electrumx.GetScriptHistory(ctx, scriptHash)
	if err != nil {
		log.ErrorWithContextf(ctx, "获取脚本历史记录失败: %v", err)
		return nil, err
	}

	// 转换类型为[]map[string]interface{}以便于处理
	historyArray := make([]map[string]interface{}, 0, len(scriptHistory))
	for _, item := range scriptHistory {
		historyArray = append(historyArray, map[string]interface{}{
			"tx_hash": item.TxHash,
			"height":  item.Height,
		})
	}

	if len(historyArray) == 0 {
		return []map[string]interface{}{}, nil
	}

	// 反转历史记录列表
	for i, j := 0, len(historyArray)-1; i < j; i, j = i+1, j-1 {
		historyArray[i], historyArray[j] = historyArray[j], historyArray[i]
	}

	// 分页处理
	start := page * size
	end := (page + 1) * size
	if start >= len(historyArray) {
		log.InfoWithContextf(ctx, "请求的页码超出历史记录范围")
		return []map[string]interface{}{}, nil
	}
	if end > len(historyArray) {
		end = len(historyArray)
	}

	return historyArray[start:end], nil
}

// processHistoryTransaction 处理历史交易
func (l *FtLogic) processHistoryTransaction(ctx context.Context, txHash string, poolId string) (ft.TBC20PoolHistoryResponse, error) {
	// 获取历史交易详情
	historyDecodeTx, err := blockchain.DecodeRawTransaction(ctx, txHash)
	if err != nil {
		return ft.TBC20PoolHistoryResponse{}, fmt.Errorf("获取历史交易详情失败: %v", err)
	}

	historyTxMap, ok := historyDecodeTx.(map[string]interface{})
	if !ok {
		return ft.TBC20PoolHistoryResponse{}, fmt.Errorf("历史交易数据类型转换失败")
	}

	// 获取交易数据
	vinArray, _ := historyTxMap["vin"].([]interface{})
	voutArray, _ := historyTxMap["vout"].([]interface{})

	// 获取交换地址
	exchangeAddress := l.getExchangeAddress(ctx, vinArray)

	// 获取上一次池子余额和当前池子余额
	lastFtLpBalance, lastFtABalance, lastTbcBalance := l.getLastPoolBalance(ctx, vinArray)
	ftLpBalance, ftABalance, tbcBalance := l.getCurrentPoolBalance(ctx, voutArray)

	// 计算余额变化
	ftLpBalanceChange := ftLpBalance - lastFtLpBalance
	ftABalanceChange := ftABalance - lastFtABalance
	tbcBalanceChange := tbcBalance - lastTbcBalance

	// 获取代币信息
	ftAContractId, ftAName, ftADecimal := l.getTokenInfo(ctx, voutArray)

	// 构造池子历史记录响应
	return ft.TBC20PoolHistoryResponse{
		Txid:                        txHash,
		PoolId:                      poolId,
		ExchangeAddress:             exchangeAddress,
		FtLpBalanceChange:           formatBalanceChange(ftLpBalanceChange),
		TokenPairAId:                "TBC",
		TokenPairAName:              "TBC",
		TokenPairADecimal:           6, // TBC默认精度
		TokenPairAPoolBalanceChange: formatBalanceChange(tbcBalanceChange),
		TokenPairBId:                ftAContractId,
		TokenPairBName:              ftAName,
		TokenPairBDecimal:           ftADecimal,
		TokenPairBPoolBalanceChange: formatBalanceChange(ftABalanceChange),
	}, nil
}

// getExchangeAddress 从交易输入中获取交换地址
func (l *FtLogic) getExchangeAddress(ctx context.Context, vinArray []interface{}) string {
	if len(vinArray) == 0 {
		return ""
	}

	for _, vinInterface := range vinArray {
		vin, ok := vinInterface.(map[string]interface{})
		if !ok {
			continue
		}

		scriptSig, ok := vin["scriptSig"].(map[string]interface{})
		if !ok {
			continue
		}

		asm, ok := scriptSig["asm"].(string)
		if !ok {
			continue
		}

		// 检查脚本特征
		if len(asm) > 0 && asm[0:2] == "30" && len(asm) < 500 {
			// 从公钥获取地址
			if len(asm) >= 66 {
				pubKey := asm[len(asm)-66:]
				// 临时使用AddressToPublicKeyHash代替
				pubKeyHash, err := utility.ConvertAddressToPublicKeyHash(pubKey)
				if err == nil {
					return pubKeyHash
				}
				log.WarnWithContextf(ctx, "公钥转换为地址失败: %v", err)
			}
		}
	}

	return ""
}

// getLastPoolBalance 获取上一次池子余额
func (l *FtLogic) getLastPoolBalance(ctx context.Context, vinArray []interface{}) (int64, int64, int64) {
	if len(vinArray) == 0 {
		return 0, 0, 0
	}

	firstVin, ok := vinArray[0].(map[string]interface{})
	if !ok {
		return 0, 0, 0
	}

	scriptSig, ok := firstVin["scriptSig"].(map[string]interface{})
	if !ok {
		return 0, 0, 0
	}

	asm, ok := scriptSig["asm"].(string)
	if !ok || len(asm) <= 0 || asm[0:2] != "30" || len(asm) <= 500 {
		return 0, 0, 0
	}

	lastTxid, ok := firstVin["txid"].(string)
	if !ok {
		return 0, 0, 0
	}

	decodedLastTx, err := blockchain.DecodeRawTransaction(ctx, lastTxid)
	if err != nil {
		log.WarnWithContextf(ctx, "获取上一笔交易失败: %v", err)
		return 0, 0, 0
	}

	decodedLastTxMap, ok := decodedLastTx.(map[string]interface{})
	if !ok {
		return 0, 0, 0
	}

	lastVoutArray, ok := decodedLastTxMap["vout"].([]interface{})
	if !ok || len(lastVoutArray) <= 1 {
		return 0, 0, 0
	}

	lastVout, ok := lastVoutArray[1].(map[string]interface{})
	if !ok {
		return 0, 0, 0
	}

	lastScriptPubKey, ok := lastVout["scriptPubKey"].(map[string]interface{})
	if !ok {
		return 0, 0, 0
	}

	lastAsm, ok := lastScriptPubKey["asm"].(string)
	if !ok {
		return 0, 0, 0
	}

	lastFtLpBalance, lastFtABalance, lastTbcBalance, err := utility.GetPoolBalanceFromTapeASM(lastAsm)
	if err != nil {
		log.WarnWithContextf(ctx, "解析上一次池子余额失败: %v", err)
		return 0, 0, 0
	}

	return lastFtLpBalance, lastFtABalance, lastTbcBalance
}

// getCurrentPoolBalance 获取当前池子余额
func (l *FtLogic) getCurrentPoolBalance(ctx context.Context, voutArray []interface{}) (int64, int64, int64) {
	if len(voutArray) <= 1 {
		return 0, 0, 0
	}

	vout, ok := voutArray[1].(map[string]interface{})
	if !ok {
		return 0, 0, 0
	}

	scriptPubKey, ok := vout["scriptPubKey"].(map[string]interface{})
	if !ok {
		return 0, 0, 0
	}

	asm, ok := scriptPubKey["asm"].(string)
	if !ok {
		return 0, 0, 0
	}

	ftLpBalance, ftABalance, tbcBalance, err := utility.GetPoolBalanceFromTapeASM(asm)
	if err != nil {
		log.WarnWithContextf(ctx, "解析当前池子余额失败: %v", err)
		return 0, 0, 0
	}

	return ftLpBalance, ftABalance, tbcBalance
}

// getTokenInfo 获取代币信息
func (l *FtLogic) getTokenInfo(ctx context.Context, voutArray []interface{}) (string, string, int) {
	var ftAContractId string
	var ftAName string
	var ftADecimal int = 8 // 默认精度

	if len(voutArray) <= 1 {
		return ftAContractId, ftAName, ftADecimal
	}

	vout1, ok := voutArray[1].(map[string]interface{})
	if !ok {
		return ftAContractId, ftAName, ftADecimal
	}

	scriptPubKey, ok := vout1["scriptPubKey"].(map[string]interface{})
	if !ok {
		return ftAContractId, ftAName, ftADecimal
	}

	asm, ok := scriptPubKey["asm"].(string)
	if !ok {
		return ftAContractId, ftAName, ftADecimal
	}

	// 解析ASM字符串，获取合约ID
	asmParts := strings.Split(asm, " ")
	if len(asmParts) <= 4 {
		return ftAContractId, ftAName, ftADecimal
	}

	ftAContractId = asmParts[4]
	// 查询数据库获取代币信息
	token, err := l.ftTokensDAO.GetFtTokenById(ftAContractId)
	if err == nil && token != nil {
		ftAName = token.FtName
		ftADecimal = int(token.FtDecimal)
	} else {
		log.WarnWithContextf(ctx, "获取代币信息失败: %v", err)
		ftAName = ftAContractId // 如果获取失败，使用合约ID作为名称
	}

	return ftAContractId, ftAName, ftADecimal
}

// formatBalanceChange 格式化余额变化为字符串，添加正负号
func formatBalanceChange(balance int64) string {
	if balance > 0 {
		return "+" + strconv.FormatInt(balance, 10)
	}
	return strconv.FormatInt(balance, 10)
}
