package ft

import (
	"context"
	"fmt"
	"strings"

	"ginproject/entity/blockchain"
	"ginproject/entity/ft"
	"ginproject/middleware/log"
	rpcblockchain "ginproject/repo/rpc/blockchain"
)

// GetNFTPoolInfoByContractId 根据代币合约ID获取NFT池信息
func (l *FtLogic) GetNFTPoolInfoByContractId(ctx context.Context, req *ft.TBC20PoolNFTInfoRequest) (*ft.TBC20PoolNFTInfoResponse, error) {
	// 使用entity层的验证逻辑
	if err := req.Validate(); err != nil {
		log.ErrorWithContextf(ctx, "参数验证失败: %v", err)
		return nil, fmt.Errorf("参数验证失败: %w", err)
	}

	// 1. 查询NFT池交易ID和余额
	log.InfoWithContextf(ctx, "开始查询NFT池信息: 合约ID=%s", req.FtContractId)
	nftUtxoId, nftCodeBalance, err := l.ftPoolNftDAO.GetPoolNftInfoByContractId(ctx, req.FtContractId)
	if err != nil {
		log.ErrorWithContextf(ctx, "查询NFT池信息失败: %v", err)
		return nil, fmt.Errorf("查询NFT池信息失败: %w", err)
	}

	// 检查是否找到NFT池
	if nftUtxoId == "" {
		log.ErrorWithContextf(ctx, "未找到NFT池: 合约ID=%s", req.FtContractId)
		return nil, fmt.Errorf("no pool NFT found")
	}

	// 2. 解码交易信息
	log.InfoWithContextf(ctx, "开始解码NFT池交易: 交易ID=%s", nftUtxoId)
	decodeTx, err := l.DecodeTxHash(ctx, nftUtxoId)
	if err != nil {
		log.ErrorWithContextf(ctx, "解码NFT池交易失败: %v", err)
		return nil, fmt.Errorf("decode pool NFT failed")
	}

	// 3. 初始化响应对象
	response := &ft.TBC20PoolNFTInfoResponse{}
	currentPoolNftTxid := nftUtxoId
	currentPoolNftBalance := int64(nftCodeBalance)
	var currentPoolNftVout int64 = 0

	// 设置基本的NFT池信息
	response.CurrentPoolNftTxid = &currentPoolNftTxid
	response.CurrentPoolNftVout = &currentPoolNftVout
	response.CurrentPoolNftBalance = &currentPoolNftBalance

	// 4. 提取和解析池服务提供商信息
	poolServiceProvider, poolVersion := l.extractPoolProviderInfo(decodeTx)
	response.PoolServiceProvider = &poolServiceProvider
	response.PoolVersion = &poolVersion

	// 检查交易输出数量
	if len(decodeTx.Vout) < 2 {
		log.ErrorWithContextf(ctx, "交易输出数量不足: 交易ID=%s, 输出数量=%d", nftUtxoId, len(decodeTx.Vout))
		return nil, fmt.Errorf("decode pool NFT failed")
	}

	// 5. 解析磁带脚本并提取余额信息
	err = l.parseTapeScriptAndSetResponse(ctx, decodeTx, response)
	if err != nil {
		return nil, err
	}

	// 6. 记录成功日志
	log.InfoWithContextf(ctx, "NFT池信息查询成功: 合约ID=%s, 交易ID=%s, 余额=%d, 服务提供商=%s, 版本=%d",
		req.FtContractId, currentPoolNftTxid, currentPoolNftBalance, poolServiceProvider, *response.PoolVersion)

	return response, nil
}

// extractPoolProviderInfo 提取池服务提供商信息
func (l *FtLogic) extractPoolProviderInfo(decodeTx *blockchain.TransactionResponse) (string, int64) {
	var poolServiceProviderHex string = ""
	var poolVersion int64 = 1

	// 检查vout数组长度以避免索引越界
	if len(decodeTx.Vout) > 0 {
		asmParts := strings.Split(decodeTx.Vout[0].ScriptPubKey.Asm, " ")
		if len(asmParts) >= 2 && asmParts[len(asmParts)-2] != "OP_RETURN" {
			poolServiceProviderHex = asmParts[len(asmParts)-2]
			poolVersion = 2
		}
	}

	// 尝试将服务提供商十六进制转换为字符串
	poolServiceProvider := ""
	if poolServiceProviderHex != "" {
		hexValue := blockchain.DigitalAsmToHex(poolServiceProviderHex)
		serviceProviderBytes, err := blockchain.HexToString(hexValue)
		if err == nil {
			poolServiceProvider = serviceProviderBytes
		}
	}

	return poolServiceProvider, poolVersion
}

// parseTapeScriptAndSetResponse 解析磁带脚本并设置响应值
func (l *FtLogic) parseTapeScriptAndSetResponse(ctx context.Context, decodeTx *blockchain.TransactionResponse, response *ft.TBC20PoolNFTInfoResponse) error {
	// 解析磁带脚本
	tapeScriptAsm := decodeTx.Vout[1].ScriptPubKey.Asm
	tapeScriptAsmList := strings.Split(tapeScriptAsm, " ")

	// 检查磁带脚本格式是否正确
	if len(tapeScriptAsmList) < 6 {
		log.ErrorWithContextf(ctx, "磁带脚本格式不正确: 元素数量=%d", len(tapeScriptAsmList))
		return fmt.Errorf("decode pool NFT failed")
	}

	// 获取服务费率
	var poolServiceFeeRateStr *string = nil
	if len(tapeScriptAsmList) == 7 {
		feeRateHex := tapeScriptAsmList[5]
		poolServiceFeeRateStr = &feeRateHex
		log.InfoWithContextf(ctx, "获取服务费率: %s", feeRateHex)
	} else {
		log.InfoWithContextf(ctx, "磁带脚本长度为%d，服务费率设为nil", len(tapeScriptAsmList))
	}
	response.PoolServiceFeeRate = poolServiceFeeRateStr

	// 获取复杂部分哈希和余额
	complexPartialHash := tapeScriptAsmList[2]
	complexBalance := tapeScriptAsmList[3]

	// 解析LP、A代币和TBC余额
	ftLpBalance, ftABalance, tbcBalance, err := l.parseBalances(ctx, complexBalance)
	if err != nil {
		return err
	}

	// 设置余额
	response.FtLpBalance = &ftLpBalance
	response.FtABalance = &ftABalance
	response.TbcBalance = &tbcBalance

	// 获取LP和A代币的部分哈希
	ftLpPartialHash := complexPartialHash[0:64]
	ftAPartialHash := complexPartialHash[64:128]
	response.FtLpPartialHash = &ftLpPartialHash
	response.FtAPartialHash = &ftAPartialHash

	// 获取A代币合约交易ID和池NFT代码脚本
	ftAContractTxid := tapeScriptAsmList[4]
	poolNftCodeScript := decodeTx.Vout[0].ScriptPubKey.Hex
	response.FtAContractTxid = &ftAContractTxid
	response.PoolNftCodeScript = &poolNftCodeScript

	return nil
}

// parseBalances 解析余额信息
func (l *FtLogic) parseBalances(ctx context.Context, complexBalance string) (int64, int64, int64, error) {
	// 解析LP余额
	ftLpBalance, err := blockchain.HexToInt64(blockchain.ReverseHexString(complexBalance, 0, 16))
	if err != nil {
		log.ErrorWithContextf(ctx, "解析LP余额失败: %v", err)
		return 0, 0, 0, fmt.Errorf("decode pool NFT failed")
	}

	// 解析A代币余额
	ftABalance, err := blockchain.HexToInt64(blockchain.ReverseHexString(complexBalance, 16, 32))
	if err != nil {
		log.ErrorWithContextf(ctx, "解析A代币余额失败: %v", err)
		return 0, 0, 0, fmt.Errorf("decode pool NFT failed")
	}

	// 解析TBC余额
	tbcBalance, err := blockchain.HexToInt64(blockchain.ReverseHexString(complexBalance, 32, 48))
	if err != nil {
		log.ErrorWithContextf(ctx, "解析TBC余额失败: %v", err)
		return 0, 0, 0, fmt.Errorf("decode pool NFT failed")
	}

	return ftLpBalance, ftABalance, tbcBalance, nil
}

// DecodeTxHash 解码交易哈希，获取交易详情
func (l *FtLogic) DecodeTxHash(ctx context.Context, txid string) (*blockchain.TransactionResponse, error) {
	if txid == "" {
		return nil, fmt.Errorf("交易ID不能为空")
	}

	log.InfoWithContextf(ctx, "开始解码交易: %s", txid)

	// 调用repo层的RPC方法解码交易
	resultChan := rpcblockchain.DecodeTxHash(ctx, txid)
	result := <-resultChan
	if result.Error != nil {
		log.ErrorWithContextf(ctx, "解码交易失败: %v", result.Error)
		return nil, fmt.Errorf("解码交易失败: %w", result.Error)
	}

	// 类型断言转换为TransactionResponse
	txInfo, ok := result.Result.(*blockchain.TransactionResponse)
	if !ok {
		log.ErrorWithContextf(ctx, "解码交易结果类型转换失败")
		return nil, fmt.Errorf("解码交易结果类型转换失败")
	}

	log.InfoWithContextf(ctx, "解码交易成功: %s", txid)
	return txInfo, nil
}
