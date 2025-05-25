package ft

import (
	"context"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"

	"ginproject/entity/ft"
	"ginproject/entity/utility"
	"ginproject/middleware/log"
	"ginproject/repo/db/ft_txo_dao"
	"ginproject/repo/db/nft_utxo_set_dao"
	"ginproject/repo/rpc/blockchain"
	"ginproject/repo/rpc/electrumx"
	blockchianEntity "ginproject/entity/blockchain"
)

// GetNFTPoolInfoByContractId 根据合约ID获取NFT池信息
func (l *FtLogic) GetNFTPoolInfoByContractId(ctx context.Context, req *ft.TBC20PoolNFTInfoRequest) (*ft.TBC20PoolNFTInfoResponse, error) {
	log.InfoWithContextf(ctx, "开始获取NFT池信息: ftContractId=%s", req.FtContractId)

	// 1. 从数据库获取NFT池信息
	nftUtxoSetDAO := nft_utxo_set_dao.NewNftUtxoSetDAO()
	currentPoolNftTxid, currentPoolNftBalance, err := nftUtxoSetDAO.GetPoolNftInfoByContractId(ctx, req.FtContractId)
	if err != nil {
		log.ErrorWithContextf(ctx, "获取NFT池信息失败: %v", err)
		return nil, fmt.Errorf("获取NFT池信息失败: %w", err)
	}

	// 2. 如果找不到NFT池，返回错误
	if currentPoolNftTxid == "" {
		log.ErrorWithContextf(ctx, "未找到NFT池: ftContractId=%s", req.FtContractId)
		return nil, fmt.Errorf("未找到NFT池")
	}

	// 3. 获取交易详情
	decodeTxResultChan := blockchain.DecodeTxHash(ctx, currentPoolNftTxid)
	decodeTxResult := <-decodeTxResultChan
	if decodeTxResult.Error != nil {
		log.ErrorWithContextf(ctx, "解码交易失败: %v", decodeTxResult.Error)
		return nil, fmt.Errorf("解码交易失败: %w", decodeTxResult.Error)
	}

	decodeTx, ok := decodeTxResult.Result.(*blockchianEntity.TransactionResponse)
	if !ok {
		log.ErrorWithContextf(ctx, "解码交易结果类型错误: txid=%s", currentPoolNftTxid)
		return nil, fmt.Errorf("解码交易结果类型错误")
	}

	// 4. 解析交易输出
	vout := decodeTx.Vout
	if len(vout) < 2 {
		log.ErrorWithContextf(ctx, "解码交易输出错误: txid=%s", currentPoolNftTxid)
		return nil, fmt.Errorf("解码交易输出错误")
	}

	// 5. 获取池服务提供商和版本
	vout0 := vout[0]

	scriptPubKey0 := vout0.ScriptPubKey

	asm0 := scriptPubKey0.Asm
	if asm0 == "" {
		log.ErrorWithContextf(ctx, "解析asm失败: txid=%s", currentPoolNftTxid)
		return nil, fmt.Errorf("解析asm失败")
	}

	asmParts0 := strings.Split(asm0, " ")
	poolVersion := int64(1)
	poolServiceProviderHex := ""

	if len(asmParts0) >= 2 && asmParts0[len(asmParts0)-2] != "OP_RETURN" {
		poolServiceProviderHex = asmParts0[len(asmParts0)-2]
		poolVersion = 2
	}

	poolServiceProvider := ""
	if poolServiceProviderHex != "" {
		digitalHex := utility.DigitalAsmToHex(poolServiceProviderHex)
		decodedBytes, err := hex.DecodeString(digitalHex)
		if err == nil {
			poolServiceProvider = string(decodedBytes)
		}
	}

	// 6. 解析交易输出1（复杂余额和部分哈希）
	vout1 := vout[1]

	scriptPubKey1 := vout1.ScriptPubKey

	asm1 := scriptPubKey1.Asm
	if asm1 == "" {
		log.ErrorWithContextf(ctx, "解析scriptPubKey1失败: txid=%s", currentPoolNftTxid)
		return nil, fmt.Errorf("解析scriptPubKey1失败")
	}

	tapeAsmList := strings.Split(asm1, " ")
	if len(tapeAsmList) < 6 {
		log.ErrorWithContextf(ctx, "解析交易脚本失败，数据不完整: txid=%s", currentPoolNftTxid)
		return nil, fmt.Errorf("解析交易脚本失败，数据不完整")
	}

	var poolServiceFeeRate *int
	if len(tapeAsmList) >= 7 {
		feeRateStr := tapeAsmList[5]
		feeRate, err := strconv.Atoi(feeRateStr)
		if err == nil {
			poolServiceFeeRate = &feeRate
		}
	}

	complexPartialHash := tapeAsmList[2]
	complexBalance := tapeAsmList[3]
	ftAContractTxid := tapeAsmList[4]

	// 7. 解析复杂余额
	ftLpBalance, ftABalance, tbcBalance, err := parsePoolBalance(complexBalance)
	if err != nil {
		log.ErrorWithContextf(ctx, "解析池余额失败: %v", err)
		return nil, fmt.Errorf("解析池余额失败: %w", err)
	}

	// 8. 解析部分哈希
	ftLpPartialHash := complexPartialHash[0:64]
	ftAPartialHash := complexPartialHash[64:128]

	// 9. 获取池NFT代码脚本
	poolNftCodeScript := scriptPubKey0.Hex
	if poolNftCodeScript == "" {
		log.ErrorWithContextf(ctx, "获取池NFT代码脚本失败: txid=%s", currentPoolNftTxid)
		return nil, fmt.Errorf("获取池NFT代码脚本失败")
	}

	// 10. 构建响应
	currentPoolNftBalanceInt64 := int64(currentPoolNftBalance)
	defaultVout := int64(0)

	poolNftInfo := &ft.TBC20PoolNFTInfoResponse{
		FtLpBalance:           &ftLpBalance,
		FtABalance:            &ftABalance,
		TbcBalance:            &tbcBalance,
		PoolVersion:           &poolVersion,
		PoolServiceFeeRate:    poolServiceFeeRate,
		PoolServiceProvider:   &poolServiceProvider,
		FtLpPartialHash:       &ftLpPartialHash,
		FtAPartialHash:        &ftAPartialHash,
		FtAContractTxid:       &ftAContractTxid,
		PoolNftCodeScript:     &poolNftCodeScript,
		CurrentPoolNftTxid:    &currentPoolNftTxid,
		CurrentPoolNftVout:    &defaultVout,
		CurrentPoolNftBalance: &currentPoolNftBalanceInt64,
	}

	log.InfoWithContextf(ctx, "成功获取NFT池信息: ftContractId=%s", req.FtContractId)
	return poolNftInfo, nil
}

// parsePoolBalance 解析池余额
func parsePoolBalance(complexBalance string) (int64, int64, int64, error) {
	if len(complexBalance) < 48 {
		return 0, 0, 0, fmt.Errorf("复杂余额格式错误，长度不足")
	}

	// 将16个字节的LP余额反转并转换为整数
	ftLpBalanceStr := ""
	for i := 14; i >= 0; i -= 2 {
		ftLpBalanceStr += complexBalance[i : i+2]
	}
	ftLpBalance, err := strconv.ParseInt(ftLpBalanceStr, 16, 64)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("解析LP余额失败: %w", err)
	}

	// 将16个字节的代币A余额反转并转换为整数
	ftABalanceStr := ""
	for i := 30; i >= 16; i -= 2 {
		ftABalanceStr += complexBalance[i : i+2]
	}
	ftABalance, err := strconv.ParseInt(ftABalanceStr, 16, 64)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("解析代币A余额失败: %w", err)
	}

	// 将16个字节的TBC余额反转并转换为整数
	tbcBalanceStr := ""
	for i := 46; i >= 32; i -= 2 {
		tbcBalanceStr += complexBalance[i : i+2]
	}
	tbcBalance, err := strconv.ParseInt(tbcBalanceStr, 16, 64)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("解析TBC余额失败: %w", err)
	}

	return ftLpBalance, ftABalance, tbcBalance, nil
}

// GetLPUnspentByScriptHash 根据脚本哈希获取LP未花费交易输出
func (l *FtLogic) GetLPUnspentByScriptHash(ctx context.Context, req *ft.LPUnspentByScriptHashRequest) (*ft.TBC20FTLPUnspentResponse, error) {
	log.InfoWithContext(ctx, "开始获取LP未花费交易输出", "scriptHash", req.ScriptHash)

	// 1. 从ElectrumX获取未花费的交易输出
	unspents, err := electrumx.GetScriptHashUnspent(ctx, req.ScriptHash)
	if err != nil {
		log.ErrorWithContext(ctx, "获取未花费交易输出失败", "scriptHash", req.ScriptHash, "error", err)
		return nil, err
	}

	// 2. 如果没有未花费的交易输出，返回空列表
	if len(unspents) == 0 {
		log.InfoWithContext(ctx, "未找到未花费交易输出", "scriptHash", req.ScriptHash)
		return &ft.TBC20FTLPUnspentResponse{FtUtxoList: []*ft.TBC20FTLPUnspentItem{}}, nil
	}

	// 3. 准备UTXO ID和输出索引列表
	txids := make([]string, 0, len(unspents))
	vouts := make([]int, 0, len(unspents))
	for _, utxo := range unspents {
		txids = append(txids, utxo.TxHash)
		vouts = append(vouts, utxo.TxPos)
	}

	// 4. 获取代币交易输出信息
	ftTxoDAO := ft_txo_dao.NewFtTxoDAO()
	ftTxos, err := ftTxoDAO.GetLPUnspentByIds(ctx, txids, vouts)
	if err != nil {
		log.ErrorWithContext(ctx, "查询代币交易输出失败", "scriptHash", req.ScriptHash, "error", err)
		return nil, err
	}

	// 5. 构建响应
	ftUtxoList := make([]*ft.TBC20FTLPUnspentItem, 0, len(ftTxos))
	for _, ftTxo := range ftTxos {
		ftUtxoList = append(ftUtxoList, &ft.TBC20FTLPUnspentItem{
			UtxoId:       ftTxo.UtxoTxid,
			UtxoVout:     ftTxo.UtxoVout,
			UtxoBalance:  ftTxo.UtxoBalance,
			FtContractId: ftTxo.FtContractId,
			FtBalance:    int64(ftTxo.FtBalance),
		})
	}

	log.InfoWithContextf(ctx, "成功获取LP未花费交易输出: scriptHash=%s, count=%d", req.ScriptHash, len(ftUtxoList))
	return &ft.TBC20FTLPUnspentResponse{FtUtxoList: ftUtxoList}, nil
}
