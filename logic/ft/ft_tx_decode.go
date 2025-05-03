package ft

import (
	"context"
	"fmt"

	"ginproject/entity/ft"
	"ginproject/entity/utility"
	"ginproject/middleware/log"
	rpcblockchain "ginproject/repo/rpc/blockchain"
)

// DecodeFtTransactionHistory 解析FT交易历史
func (l *FtLogic) DecodeFtTransactionHistory(ctx context.Context, req *ft.FtTxDecodeRequest) (*ft.FtTxDecodeResponse, error) {
	// 验证请求参数
	if err := ft.ValidateFtTxDecodeRequest(req); err != nil {
		log.ErrorWithContextf(ctx, "验证交易参数失败: %v", err)
		return nil, fmt.Errorf("参数验证失败: %v", err)
	}

	// 记录操作日志
	log.InfoWithContextf(ctx, "开始解析FT交易: 交易ID=%s", req.Txid)

	// 调用RPC解析交易
	decode_tx, err := rpcblockchain.CallRPC("getrawtransaction", []interface{}{req.Txid, 1}, false)
	if err != nil {
		log.ErrorWithContextf(ctx, "解析交易失败: %v", err)
		return nil, fmt.Errorf("解析交易失败: %v", err)
	}

	// 准备返回数据
	input_list := []ft.FtTxDecodeData{}
	output_list := []ft.FtTxDecodeData{}

	// 提取交易输入信息
	decodeTxMap, ok := decode_tx.(map[string]interface{})
	if !ok {
		log.ErrorWithContextf(ctx, "交易数据格式错误")
		return nil, fmt.Errorf("交易数据格式错误")
	}

	// 处理交易输入
	if vinArray, ok := decodeTxMap["vin"].([]interface{}); ok {
		for vinIndex, vin := range vinArray {
			vinMap, ok := vin.(map[string]interface{})
			if !ok {
				continue
			}

			// 获取输入交易ID和输出索引
			inputTxid, _ := vinMap["txid"].(string)
			vout, _ := vinMap["vout"].(float64)

			if inputTxid == "" {
				continue
			}

			// 查询ft_txo_set表获取FT代币信息
			ftBalance, ftHolderScript, ftContractId, err := l.ftTxoDAO.GetFtUtxoInfo(ctx, inputTxid, int(vout))
			if err != nil {
				log.WarnWithContextf(ctx, "获取FT UTXO信息失败: txid=%s, vout=%d, 错误=%v",
					inputTxid, int(vout), err)
				continue
			}

			// 如果不是FT代币相关交易，则跳过
			if ftHolderScript == "" || ftContractId == "" {
				continue
			}

			// 获取代币小数位数
			ftDecimal, err := l.ftTokensDAO.GetFtDecimalByContractId(ctx, ftContractId)
			if err != nil {
				log.WarnWithContextf(ctx, "获取代币小数位数失败: 合约ID=%s, 错误=%v",
					ftContractId, err)
				ftDecimal = 0 // 默认值
			}

			// 处理地址转换
			address := ""
			if ftHolderScript[len(ftHolderScript)-2:] == "00" {
				// 普通地址
				address, err = utility.ConvertCombineScriptToAddress(ftHolderScript)
				if err != nil {
					log.WarnWithContextf(ctx, "转换组合脚本为地址失败: %v", err)
				}
			} else if ftHolderScript[len(ftHolderScript)-2:] == "01" {
				// 多签地址或池控制地址
				scriptSig, ok := vinMap["scriptSig"].(map[string]interface{})
				if ok {
					asm, ok := scriptSig["asm"].(string)
					if ok && vinIndex > 0 && len(asm) > 2 && asm[:2] == "0 " {
						// 多签地址
						address, err = utility.ConvertP2msUnlockScriptToAddress(asm)
						if err != nil {
							log.WarnWithContextf(ctx, "转换P2MS解锁脚本为地址失败: %v", err)
							address = "Pool_" + ftHolderScript
						}
					} else {
						// 池控制地址
						address = "Pool_" + ftHolderScript
					}
				}
			}

			// 创建输入项
			inputItem := ft.FtTxDecodeData{
				Txid:       inputTxid,
				Vout:       int(vout),
				Address:    address,
				ContractId: ftContractId,
				FtBalance:  int64(ftBalance),
				FtDecimal:  int(ftDecimal),
			}
			input_list = append(input_list, inputItem)

			log.DebugWithContextf(ctx, "添加输入项: txid=%s, vout=%d, 地址=%s, 合约ID=%s, 余额=%d",
				inputTxid, int(vout), address, ftContractId, ftBalance)
		}
	}

	// 处理交易输出
	if voutArray, ok := decodeTxMap["vout"].([]interface{}); ok {
		for _, vout := range voutArray {
			voutMap, ok := vout.(map[string]interface{})
			if !ok {
				continue
			}

			// 获取输出索引
			n, _ := voutMap["n"].(float64)

			// 查询ft_txo_set表获取FT代币信息
			ftBalance, ftHolderScript, ftContractId, err := l.ftTxoDAO.GetFtUtxoInfo(ctx, req.Txid, int(n))
			if err != nil {
				log.WarnWithContextf(ctx, "获取FT UTXO信息失败: txid=%s, vout=%d, 错误=%v",
					req.Txid, int(n), err)
				continue
			}

			// 如果不是FT代币相关交易，则跳过
			if ftHolderScript == "" || ftContractId == "" {
				continue
			}

			// 获取代币小数位数
			ftDecimal, err := l.ftTokensDAO.GetFtDecimalByContractId(ctx, ftContractId)
			if err != nil {
				log.WarnWithContextf(ctx, "获取代币小数位数失败: 合约ID=%s, 错误=%v",
					ftContractId, err)
				ftDecimal = 0 // 默认值
			}

			// 处理地址转换
			address := ""
			if ftHolderScript[len(ftHolderScript)-2:] == "00" {
				// 普通地址
				address, err = utility.ConvertCombineScriptToAddress(ftHolderScript)
				if err != nil {
					log.WarnWithContextf(ctx, "转换组合脚本为地址失败: %v", err)
				}
			} else if ftHolderScript[len(ftHolderScript)-2:] == "01" {
				// 池控制或多签地址
				isFound := false
				for _, input := range input_list {
					if input.Address == "Pool_"+ftHolderScript {
						address = "Pool_" + ftHolderScript
						isFound = true
						break
					}
				}
				if !isFound {
					address = "Pool_or_MS_" + ftHolderScript
				}
			}

			// 创建输出项
			outputItem := ft.FtTxDecodeData{
				Txid:       req.Txid,
				Vout:       int(n),
				Address:    address,
				ContractId: ftContractId,
				FtBalance:  int64(ftBalance),
				FtDecimal:  int(ftDecimal),
			}
			output_list = append(output_list, outputItem)

			log.DebugWithContextf(ctx, "添加输出项: txid=%s, vout=%d, 地址=%s, 合约ID=%s, 余额=%d",
				req.Txid, int(n), address, ftContractId, ftBalance)
		}
	}

	// 构建响应
	response := &ft.FtTxDecodeResponse{
		Txid:   req.Txid,
		Input:  input_list,
		Output: output_list,
	}

	log.InfoWithContextf(ctx, "解析FT交易成功: 交易ID=%s, 输入数=%d, 输出数=%d",
		req.Txid, len(input_list), len(output_list))

	return response, nil
}
