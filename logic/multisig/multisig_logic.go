package multisig

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"ginproject/entity/multisig"
	"ginproject/entity/utility"
	"ginproject/middleware/log"
	"ginproject/repo/rpc/blockchain"
	"ginproject/repo/rpc/electrumx"
)

// 错误定义
var (
	ErrEmptyAddress    = fmt.Errorf("地址不能为空")
	ErrInvalidOpReturn = fmt.Errorf("无效的OP_RETURN数据")
	ErrNoMultisigData  = fmt.Errorf("未找到多签名数据")
)

// QueryMultiWalletByAddress 根据地址查询多签名钱包信息
// 实现与Python代码query_multi_address_and_pubkeys_by_address相同的功能
func QueryMultiWalletByAddress(ctx context.Context, address string) (*multisig.MultiWalletResponse, error) {
	// 参数校验
	if address == "" {
		log.ErrorWithContext(ctx, "地址不能为空")
		return nil, ErrEmptyAddress
	}

	log.InfoWithContext(ctx, "开始查询多签名地址", "address", address)

	// 将地址转换为多签名脚本哈希
	scriptHash, err := utility.ConvertAddressToMultiSigScriptHash(address)
	if err != nil {
		log.ErrorWithContext(ctx, "地址转换为多签名脚本哈希失败", "address", address, "error", err)
		return nil, fmt.Errorf("地址转换为多签名脚本哈希失败: %w", err)
	}

	// 获取未花费的UTXO
	unspentsChan := electrumx.CallMethodAsync(ctx, "blockchain.scripthash.listunspent", []interface{}{scriptHash})
	unspentsResult := <-unspentsChan
	if unspentsResult.Error != nil {
		log.ErrorWithContext(ctx, "获取脚本哈希UTXO失败", "scriptHash", scriptHash, "error", unspentsResult.Error)
		return nil, fmt.Errorf("获取脚本哈希UTXO失败: %w", unspentsResult.Error)
	}

	// 解析UTXO列表
	var unspents []map[string]interface{}
	if err := json.Unmarshal(unspentsResult.Result, &unspents); err != nil {
		log.ErrorWithContext(ctx, "解析UTXO列表失败", "error", err)
		return nil, fmt.Errorf("解析UTXO列表失败: %w", err)
	}

	// 构建结果列表
	multiWalletList := []multisig.MultiWallet{}

	// 遍历所有UTXO
	for _, unspent := range unspents {
		// 获取交易哈希
		txHash, ok := unspent["tx_hash"].(string)
		if !ok {
			log.WarnWithContext(ctx, "无效的交易哈希格式，跳过", "unspent", unspent)
			continue
		}

		// 调用节点RPC获取交易详情
		txChan := blockchain.CallRPCAsync(ctx, "getrawtransaction", []interface{}{txHash, 1}, false)
		txResult := <-txChan
		if txResult.Error != nil {
			log.WarnWithContext(ctx, "获取交易详情失败，跳过", "txHash", txHash, "error", txResult.Error)
			continue
		}

		// 解析交易详情
		decodeTx, ok := txResult.Result.(map[string]interface{})
		if !ok {
			log.WarnWithContext(ctx, "解析交易详情失败，跳过", "txHash", txHash)
			continue
		}

		// 获取交易输出
		vouts, ok := decodeTx["vout"].([]interface{})
		if !ok {
			log.WarnWithContext(ctx, "交易输出格式错误，跳过", "txHash", txHash)
			continue
		}

		// 遍历交易输出
		for _, vout := range vouts {
			voutMap, ok := vout.(map[string]interface{})
			if !ok {
				continue
			}

			scriptPubKey, ok := voutMap["scriptPubKey"].(map[string]interface{})
			if !ok {
				continue
			}

			asm, ok := scriptPubKey["asm"].(string)
			if !ok {
				continue
			}

			// 检查是否是以"0 OP_RETURN"开头的输出
			if !strings.HasPrefix(asm, "0 OP_RETURN") {
				continue
			}

			// 提取OP_RETURN数据: 从12个字符开始，到倒数第11个字符结束
			// 与Python代码保持一致: tape_data = vout["scriptPubKey"]["asm"][12:-11]
			// 但需要确保长度足够
			if len(asm) < 23 { // 12 + 11
				log.WarnWithContext(ctx, "OP_RETURN数据太短，跳过", "asm", asm)
				continue
			}
			tapeData := asm[12 : len(asm)-11]

			// 转换为JSON
			tapeJson, err := utility.HexToJson(tapeData)
			if err != nil {
				log.WarnWithContext(ctx, "解析OP_RETURN数据失败，跳过", "data", tapeData, "error", err)
				continue
			}

			// 提取多签名地址和公钥列表
			multiAddress, _ := tapeJson["address"].(string)
			pubkeysInterface, ok := tapeJson["pubkeys"].([]interface{})
			if !ok {
				pubkeysInterface = []interface{}{}
			}

			// 转换公钥列表
			pubkeys := make([]string, 0, len(pubkeysInterface))
			for _, p := range pubkeysInterface {
				if pk, ok := p.(string); ok {
					pubkeys = append(pubkeys, pk)
				}
			}

			// 添加到结果中
			multiWalletList = append(multiWalletList, multisig.MultiWallet{
				MultiAddress: multiAddress,
				PubkeyList:   pubkeys,
			})
		}
	}

	log.InfoWithContext(ctx, "成功查询多签名地址", "address", address, "count", len(multiWalletList))

	return &multisig.MultiWalletResponse{
		MultiWalletList: multiWalletList,
	}, nil
}
