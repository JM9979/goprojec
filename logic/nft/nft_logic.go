package nft

import (
	"context"
	"encoding/binary"
	"fmt"
	"strconv"
	"time"

	"ginproject/entity/nft"
	"ginproject/entity/utility"
	"ginproject/middleware/log"
	"ginproject/repo/db/nft_collections_dao"
	"ginproject/repo/db/nft_utxo_set_dao"
	"ginproject/repo/rpc/blockchain"
	"ginproject/repo/rpc/electrumx"
)

// NFTLogic NFT业务逻辑结构体
type NFTLogic struct {
	collectionsDAO *nft_collections_dao.NftCollectionsDAO
	utxoSetDAO     *nft_utxo_set_dao.NftUtxoSetDAO
}

// NewNFTLogic 创建一个新的NFTLogic实例
func NewNFTLogic() *NFTLogic {
	return &NFTLogic{
		collectionsDAO: nft_collections_dao.NewNftCollectionsDAO(),
		utxoSetDAO:     nft_utxo_set_dao.NewNftUtxoSetDAO(),
	}
}

// GetCollectionByAddressPageSize 根据地址、页码和每页大小获取NFT集合列表
func (logic *NFTLogic) GetCollectionByAddressPageSize(ctx context.Context, address string, page, size int) (*nft.CollectionListResponse, error) {
	// 参数校验
	if err := nft.ValidateCollectionQueryByAddress(address, page, size); err != nil {
		log.ErrorWithContext(ctx, "参数校验失败:", err)
		return nil, err
	}

	log.InfoWithContextf(ctx, "开始获取地址[%s]的NFT集合列表，页码: %d, 每页大小: %d", address, page, size)

	// 从数据库获取集合数据
	collections, total, err := logic.collectionsDAO.GetCollectionsByAddressWithPagination(ctx, address, page, size)
	if err != nil {
		log.ErrorWithContextf(ctx, "获取地址[%s]的NFT集合列表失败: %v", address, err)
		return nil, fmt.Errorf("获取集合列表失败: %v", err)
	}

	// 构建响应数据
	response := &nft.CollectionListResponse{
		CollectionCount: int(total),
		CollectionList:  make([]nft.CollectionItem, 0, len(collections)),
	}

	// 转换数据格式
	for _, collection := range collections {
		collectionItem := nft.CollectionItem{
			CollectionId:              collection.CollectionId,
			CollectionName:            collection.CollectionName,
			CollectionCreator:         collection.CollectionCreatorAddress,
			CollectionSymbol:          collection.CollectionSymbol,
			CollectionAttributes:      collection.CollectionAttributes,
			CollectionDescription:     collection.CollectionDescription,
			CollectionSupply:          collection.CollectionSupply,
			CollectionCreateTimestamp: collection.CollectionCreateTimestamp,
			CollectionIcon:            collection.CollectionIcon,
		}
		response.CollectionList = append(response.CollectionList, collectionItem)
	}

	log.InfoWithContextf(ctx, "成功获取地址[%s]的NFT集合列表，共%d条记录", address, total)
	return response, nil
}

// GetNftByAddressPageSize 根据地址、页码和每页大小获取NFT列表
func (logic *NFTLogic) GetNftByAddressPageSize(ctx context.Context, address string, page, size int, ifExtraCollectionInfo bool) (*nft.NftListResponse, error) {
	// 参数校验
	if err := nft.ValidateGetNftByAddressPageSize(address, page, size, ifExtraCollectionInfo); err != nil {
		log.ErrorWithContext(ctx, "参数校验失败:", err)
		return nil, err
	}

	log.InfoWithContextf(ctx, "开始获取地址[%s]的NFT列表，页码: %d, 每页大小: %d, 是否需要额外的集合信息: %v",
		address, page, size, ifExtraCollectionInfo)

	// 将地址转换为NFT脚本哈希
	nftScriptHash, err := convertAddressToNftScriptHash(ctx, address, false)
	if err != nil {
		log.ErrorWithContextf(ctx, "将地址[%s]转换为NFT脚本哈希失败: %v", address, err)
		return nil, fmt.Errorf("地址转换失败: %v", err)
	}

	// 从数据库获取NFT数据
	nfts, total, err := logic.utxoSetDAO.GetNftsByHolderWithPagination(ctx, nftScriptHash, page, size)
	if err != nil {
		log.ErrorWithContextf(ctx, "获取地址[%s]的NFT列表失败: %v", address, err)
		return nil, fmt.Errorf("获取NFT列表失败: %v", err)
	}

	// 构建响应数据
	response := &nft.NftListResponse{
		NftTotalCount: int(total),
		NftList:       make([]nft.NftItem, 0, len(nfts)),
	}

	// 转换数据格式
	for _, nftItem := range nfts {
		var collectionIcon, collectionDescription string

		// 如果需要额外的集合信息，则查询集合图标和描述
		if ifExtraCollectionInfo && nftItem.CollectionId != "" {
			var err error
			collectionIcon, collectionDescription, err = logic.collectionsDAO.GetCollectionIconAndDescription(ctx, nftItem.CollectionId)
			if err != nil {
				log.WarnWithContextf(ctx, "获取集合[%s]的图标和描述失败: %v", nftItem.CollectionId, err)
				// 失败不中断处理，继续使用空值
			}
		}

		// 构建NFT项目数据
		item := nft.NftItem{
			CollectionId:          nftItem.CollectionId,
			CollectionIndex:       nftItem.CollectionIndex,
			CollectionName:        nftItem.CollectionName,
			CollectionIcon:        collectionIcon,
			CollectionDescription: collectionDescription,
			NftContractId:         nftItem.NftContractId,
			NftUtxoId:             nftItem.NftUtxoId,
			NftCodeBalance:        nftItem.NftCodeBalance,
			NftP2pkhBalance:       nftItem.NftP2pkhBalance,
			NftName:               nftItem.NftName,
			NftSymbol:             nftItem.NftSymbol,
			NftDescription:        nftItem.NftDescription,
			NftTransferTimeCount:  nftItem.NftTransferTimeCount,
			NftHolder:             address,
			NftCreateTimestamp:    nftItem.NftCreateTimestamp,
			NftIcon:               nftItem.NftIcon,
		}

		response.NftList = append(response.NftList, item)
	}

	log.InfoWithContextf(ctx, "成功获取地址[%s]的NFT列表，共%d条记录", address, total)
	return response, nil
}

// GetNftByScriptHashPageSize 根据脚本哈希、页码和每页大小获取NFT列表
func (logic *NFTLogic) GetNftByScriptHashPageSize(ctx context.Context, scriptHash string, page, size int) (*nft.NftListResponse, error) {
	// 参数校验
	if err := nft.ValidateGetNftByScriptHashPageSize(scriptHash, page, size); err != nil {
		log.ErrorWithContext(ctx, "参数校验失败:", err)
		return nil, err
	}

	log.InfoWithContextf(ctx, "开始获取脚本哈希[%s]的NFT列表，页码: %d, 每页大小: %d", scriptHash, page, size)

	// 获取未花费交易输出
	unspents, err := electrumx.GetUnspent(ctx, scriptHash)
	if err != nil {
		log.ErrorWithContextf(ctx, "获取脚本哈希[%s]的未花费交易失败: %v", scriptHash, err)
		return nil, fmt.Errorf("获取未花费交易失败: %v", err)
	}

	// 计算总数和分页范围
	nftTotalCount := len(unspents)
	startIndex := page * size
	endIndex := startIndex + size

	// 边界检查
	if startIndex >= nftTotalCount {
		// 如果起始索引超出范围，返回空结果
		return &nft.NftListResponse{
			NftTotalCount: nftTotalCount,
			NftList:       []nft.NftItem{},
		}, nil
	}

	// 确保结束索引不超出总数
	if endIndex > nftTotalCount {
		endIndex = nftTotalCount
	}

	// 截取当前页的交易
	pageUnspents := unspents[startIndex:endIndex]

	// 构建响应数据
	response := &nft.NftListResponse{
		NftTotalCount: nftTotalCount,
		NftList:       make([]nft.NftItem, 0, len(pageUnspents)),
	}

	// 遍历并获取详细信息
	for _, utxo := range pageUnspents {
		utxoId := utxo.TxHash

		// 查询NFT UTXO信息
		nftInfo, err := logic.utxoSetDAO.GetNftByUtxoId(ctx, utxoId)
		if err != nil {
			log.WarnWithContextf(ctx, "获取UTXO[%s]的NFT信息失败: %v", utxoId, err)
			// 跳过此条记录，继续处理下一条
			continue
		}

		if nftInfo == nil {
			log.WarnWithContextf(ctx, "未找到UTXO[%s]的NFT信息", utxoId)
			continue
		}

		// 构建NFT项目数据
		item := nft.NftItem{
			CollectionId:         nftInfo.CollectionId,
			CollectionIndex:      nftInfo.CollectionIndex,
			CollectionName:       nftInfo.CollectionName,
			NftContractId:        nftInfo.NftContractId,
			NftUtxoId:            nftInfo.NftUtxoId,
			NftCodeBalance:       nftInfo.NftCodeBalance,
			NftP2pkhBalance:      nftInfo.NftP2pkhBalance,
			NftName:              nftInfo.NftName,
			NftSymbol:            nftInfo.NftSymbol,
			NftDescription:       nftInfo.NftDescription,
			NftTransferTimeCount: nftInfo.NftTransferTimeCount,
			NftHolder:            scriptHash, // 使用脚本哈希作为持有者标识
			NftCreateTimestamp:   nftInfo.NftCreateTimestamp,
			NftIcon:              nftInfo.NftIcon,
		}

		response.NftList = append(response.NftList, item)
	}

	log.InfoWithContextf(ctx, "成功获取脚本哈希[%s]的NFT列表，共%d条记录，当前页%d条", scriptHash, nftTotalCount, len(response.NftList))
	return response, nil
}

// GetNftByCollectionIdPageSize 根据集合ID、页码和每页大小获取NFT列表
func (logic *NFTLogic) GetNftByCollectionIdPageSize(ctx context.Context, collectionId string, page, size int) (*nft.NftListResponse, error) {
	// 参数校验
	if err := nft.ValidateGetNftByCollectionIdPageSize(collectionId, page, size); err != nil {
		log.ErrorWithContext(ctx, "参数校验失败:", err)
		return nil, err
	}

	log.InfoWithContextf(ctx, "开始获取集合[%s]的NFT列表，页码: %d, 每页大小: %d", collectionId, page, size)

	// 使用DAO层方法获取数据
	nfts, total, err := logic.utxoSetDAO.GetNftsByCollectionIdWithPagination(ctx, collectionId, page, size)
	if err != nil {
		log.ErrorWithContextf(ctx, "获取集合[%s]的NFT列表失败: %v", collectionId, err)
		return nil, fmt.Errorf("获取NFT列表失败: %v", err)
	}

	// 构建响应数据
	response := &nft.NftListResponse{
		NftTotalCount: int(total),
		NftList:       make([]nft.NftItem, 0, len(nfts)),
	}

	// 转换数据格式
	for _, nftInfo := range nfts {
		// 构建NFT项目数据
		item := nft.NftItem{
			CollectionId:         nftInfo.CollectionId,
			CollectionIndex:      nftInfo.CollectionIndex,
			CollectionName:       nftInfo.CollectionName,
			NftContractId:        nftInfo.NftContractId,
			NftUtxoId:            nftInfo.NftUtxoId,
			NftCodeBalance:       nftInfo.NftCodeBalance,
			NftP2pkhBalance:      nftInfo.NftP2pkhBalance,
			NftName:              nftInfo.NftName,
			NftSymbol:            nftInfo.NftSymbol,
			NftDescription:       nftInfo.NftDescription,
			NftTransferTimeCount: nftInfo.NftTransferTimeCount,
			NftHolder:            nftInfo.NftHolderAddress,
			NftCreateTimestamp:   nftInfo.NftCreateTimestamp,
			NftIcon:              nftInfo.NftIcon,
		}

		response.NftList = append(response.NftList, item)
	}

	log.InfoWithContextf(ctx, "成功获取集合[%s]的NFT列表，共%d条记录", collectionId, total)
	return response, nil
}

// GetCollectionsByPageSize 分页获取所有NFT集合列表
func (logic *NFTLogic) GetCollectionsByPageSize(ctx context.Context, page, size int) (*nft.CollectionListResponse, error) {
	// 参数校验
	if err := nft.ValidateCollectionsPageSize(page, size); err != nil {
		log.ErrorWithContext(ctx, "参数校验失败:", err)
		return nil, err
	}

	log.InfoWithContextf(ctx, "开始获取NFT集合列表，页码: %d, 每页大小: %d", page, size)

	// 从数据库获取集合数据
	collections, total, err := logic.collectionsDAO.GetAllCollectionsWithPagination(ctx, page, size)
	if err != nil {
		log.ErrorWithContextf(ctx, "获取NFT集合列表失败: %v", err)
		return nil, fmt.Errorf("获取集合列表失败: %v", err)
	}

	// 构建响应数据
	response := &nft.CollectionListResponse{
		CollectionCount: int(total),
		CollectionList:  make([]nft.CollectionItem, 0, len(collections)),
	}

	// 转换数据格式
	for _, collection := range collections {
		collectionItem := nft.CollectionItem{
			CollectionId:              collection.CollectionId,
			CollectionName:            collection.CollectionName,
			CollectionCreator:         collection.CollectionCreatorAddress,
			CollectionSymbol:          collection.CollectionSymbol,
			CollectionAttributes:      collection.CollectionAttributes,
			CollectionDescription:     collection.CollectionDescription,
			CollectionSupply:          collection.CollectionSupply,
			CollectionCreateTimestamp: collection.CollectionCreateTimestamp,
			CollectionIcon:            collection.CollectionIcon,
		}
		response.CollectionList = append(response.CollectionList, collectionItem)
	}

	log.InfoWithContextf(ctx, "成功获取NFT集合列表，共%d条记录", total)
	return response, nil
}

// GetNftsByContractIds 根据合约ID列表获取NFT信息
func (logic *NFTLogic) GetNftsByContractIds(ctx context.Context, contractList []string, ifIconNeeded bool) (*nft.NftInfoListResponse, error) {
	// 参数校验
	if err := nft.ValidateGetNftsByContractIds(contractList, ifIconNeeded); err != nil {
		log.ErrorWithContext(ctx, "参数校验失败:", err)
		return nil, err
	}

	log.InfoWithContextf(ctx, "开始获取合约ID列表的NFT信息，合约数量: %d, 是否需要图标: %v", len(contractList), ifIconNeeded)

	// 从数据库获取NFT数据
	nfts, err := logic.utxoSetDAO.GetNftsByContractIds(ctx, contractList)
	if err != nil {
		log.ErrorWithContextf(ctx, "获取合约ID列表的NFT信息失败: %v", err)
		return nil, fmt.Errorf("获取NFT信息失败: %v", err)
	}

	// 构建响应数据
	response := &nft.NftInfoListResponse{
		NftInfoList: make([]nft.NftItem, 0, len(nfts)),
	}

	// 转换数据格式
	for _, nftInfo := range nfts {
		// 如果不需要图标，则置为空
		nftIcon := nftInfo.NftIcon
		if !ifIconNeeded {
			nftIcon = ""
		}

		// 构建NFT项目数据
		item := nft.NftItem{
			CollectionId:         nftInfo.CollectionId,
			CollectionIndex:      nftInfo.CollectionIndex,
			CollectionName:       nftInfo.CollectionName,
			NftContractId:        nftInfo.NftContractId,
			NftUtxoId:            nftInfo.NftUtxoId,
			NftCodeBalance:       nftInfo.NftCodeBalance,
			NftP2pkhBalance:      nftInfo.NftP2pkhBalance,
			NftName:              nftInfo.NftName,
			NftSymbol:            nftInfo.NftSymbol,
			NftAttributes:        nftInfo.NftAttributes,
			NftDescription:       nftInfo.NftDescription,
			NftTransferTimeCount: nftInfo.NftTransferTimeCount,
			NftHolder:            nftInfo.NftHolderAddress,
			NftCreateTimestamp:   nftInfo.NftCreateTimestamp,
			NftIcon:              nftIcon,
		}

		response.NftInfoList = append(response.NftInfoList, item)
	}

	log.InfoWithContextf(ctx, "成功获取合约ID列表的NFT信息，共%d条记录", len(response.NftInfoList))
	return response, nil
}

// GetDetailCollectionInfo 获取集合详细信息
func (logic *NFTLogic) GetDetailCollectionInfo(ctx context.Context, collectionId string) (*nft.CollectionDetailResponse, error) {
	// 参数校验
	if err := nft.ValidateDetailCollectionInfo(collectionId); err != nil {
		log.ErrorWithContext(ctx, "参数校验失败:", err)
		return nil, err
	}

	log.InfoWithContextf(ctx, "开始获取集合[%s]的详细信息", collectionId)

	// 从数据库获取集合详情
	collection, err := logic.collectionsDAO.GetDetailCollectionInfo(ctx, collectionId)
	if err != nil {
		log.ErrorWithContextf(ctx, "获取集合[%s]的详细信息失败: %v", collectionId, err)
		return nil, fmt.Errorf("获取集合详情失败: %v", err)
	}

	// 如果集合不存在，返回空结果
	if collection == nil {
		log.WarnWithContextf(ctx, "集合[%s]不存在", collectionId)
		return &nft.CollectionDetailResponse{}, nil
	}

	// 构建响应数据
	response := &nft.CollectionDetailResponse{
		CollectionId:                collection.CollectionId,
		CollectionName:              collection.CollectionName,
		CollectionCreator:           collection.CollectionCreatorAddress,
		CollectionCreatorScripthash: collection.CollectionCreatorScriptHash,
		CollectionSymbol:            collection.CollectionSymbol,
		CollectionAttributes:        collection.CollectionAttributes,
		CollectionDescription:       collection.CollectionDescription,
		CollectionSupply:            collection.CollectionSupply,
		CollectionCreateTimestamp:   collection.CollectionCreateTimestamp,
		CollectionIcon:              collection.CollectionIcon,
	}

	log.InfoWithContextf(ctx, "成功获取集合[%s]的详细信息", collectionId)
	return response, nil
}

// convertAddressToNftScriptHash 将地址转换为NFT脚本哈希
func convertAddressToNftScriptHash(ctx context.Context, address string, isCollection bool) (string, error) {
	// 使用utility包中已定义的函数进行转换
	log.DebugWithContextf(ctx, "开始将地址[%s]转换为NFT脚本哈希，isCollection=%v", address, isCollection)

	scriptHash, err := utility.ConvertAddressToNftScriptHash(address, isCollection)
	if err != nil {
		log.ErrorWithContextf(ctx, "将地址[%s]转换为NFT脚本哈希失败: %v", address, err)
		return "", err
	}

	log.DebugWithContextf(ctx, "成功将地址[%s]转换为NFT脚本哈希: %s", address, scriptHash)
	return scriptHash, nil
}

// GetNftHistoryByAddress 获取地址的NFT交易历史记录
func (logic *NFTLogic) GetNftHistoryByAddress(ctx context.Context, address string, page, size int) (*nft.NftHistoryResponse, error) {
	// 参数校验
	if err := nft.ValidateNftHistory(address, page, size); err != nil {
		log.ErrorWithContext(ctx, "参数校验失败:", err)
		return nil, err
	}

	log.InfoWithContextf(ctx, "开始获取地址[%s]的NFT历史记录，页码: %d, 每页大小: %d", address, page, size)

	// 将地址转换为NFT脚本哈希
	nftScriptHash, err := convertAddressToNftScriptHash(ctx, address, false)
	if err != nil {
		log.ErrorWithContextf(ctx, "将地址[%s]转换为NFT脚本哈希失败: %v", address, err)
		return nil, fmt.Errorf("地址转换失败: %v", err)
	}

	// 获取交易历史记录
	history, err := electrumx.GetScriptHashHistoryWithContext(ctx, nftScriptHash)
	if err != nil {
		log.ErrorWithContextf(ctx, "获取地址[%s]的交易历史记录失败: %v", address, err)
		return nil, fmt.Errorf("获取交易历史失败: %v", err)
	}

	// 反转历史记录列表，使最新的交易排在前面
	for i, j := 0, len(history)-1; i < j; i, j = i+1, j-1 {
		history[i], history[j] = history[j], history[i]
	}

	// 计算历史记录总数
	historyCount := len(history)

	// 计算分页范围
	startIndex := page * size
	endIndex := startIndex + size

	// 边界检查
	if startIndex >= historyCount {
		// 如果起始索引超出范围，返回空结果
		return &nft.NftHistoryResponse{
			Address:      address,
			ScriptHash:   nftScriptHash,
			HistoryCount: historyCount,
			Result:       []nft.NftHistoryItem{},
		}, nil
	}

	// 确保结束索引不超出总数
	if endIndex > historyCount {
		endIndex = historyCount
	}

	// 截取当前页的交易
	pageHistory := history[startIndex:endIndex]

	// 构建历史记录项目列表
	historyItems := make([]nft.NftHistoryItem, 0, len(pageHistory))

	// 处理每个交易
	for _, item := range pageHistory {
		txid := item.TxHash
		var timeStamp *int64
		var utcTime string

		// 获取交易时间戳和UTC时间
		if item.Height < 1 {
			utcTime = "未确认"
		} else {
			// 获取区块信息
			blockInfo, err := blockchain.GetBlockByHeight(ctx, int64(item.Height))
			if err != nil {
				log.WarnWithContextf(ctx, "获取区块[%d]信息失败: %v", item.Height, err)
				// 继续处理，不中断整体流程
			} else {
				// 设置时间戳和UTC时间
				timeValue, ok := blockInfo["time"].(float64)
				if ok {
					t := int64(timeValue)
					timeStamp = &t
					// 转换为UTC时间格式
					tm := time.Unix(t, 0).UTC()
					utcTime = tm.Format("2006-01-02 15:04:05")
				} else {
					log.WarnWithContextf(ctx, "解析区块[%d]时间戳失败", item.Height)
				}
			}
		}

		// 获取原始交易，解析发送者和接收者地址
		txInfo, err := blockchain.DecodeTx(ctx, txid)
		if err != nil {
			log.WarnWithContextf(ctx, "获取交易[%s]详情失败: %v", txid, err)
			// 继续处理，不中断整体流程
			continue
		}

		// 提取发送者和接收者地址
		senderAddresses := make([]string, 0)
		recipientAddresses := make([]string, 0)

		// 根据交易输入提取发送者地址
		if len(txInfo.Vin) > 0 && len(txInfo.Vin[0].ScriptSig.Hex) > 500 {
			if len(txInfo.Vin) > 1 {
				// 获取发送者公钥并转换为地址
				senderPubkey := txInfo.Vin[1].ScriptSig.Hex[len(txInfo.Vin[1].ScriptSig.Hex)-66:]
				senderAddress, err := utility.ConvertCompressedPubkeyToLegacyAddress(senderPubkey)
				if err == nil && senderAddress != "" {
					senderAddresses = append(senderAddresses, senderAddress)
				}
			}
		}

		// 根据交易输出提取接收者地址
		if len(txInfo.Vout) > 1 && len(txInfo.Vout[1].ScriptPubKey.Addresses) > 0 {
			recipientAddress := txInfo.Vout[1].ScriptPubKey.Addresses[0]
			if recipientAddress != "" {
				recipientAddresses = append(recipientAddresses, recipientAddress)
			}
		}

		// 获取NFT合约ID和索引
		var nftContractId string
		var collectionIndex int
		var tapeJson map[string]interface{}

		// 尝试从交易输出中解析NFT信息
		if len(txInfo.Vout) > 2 {
			tapeScript := txInfo.Vout[2].ScriptPubKey.Asm
			// 去除前后缀
			if len(tapeScript) > 23 {
				tapeScript = tapeScript[12 : len(tapeScript)-11]
				// 解析十六进制数据为JSON
				tapeJson, err = utility.HexToJson(tapeScript)
				if err != nil {
					log.WarnWithContextf(ctx, "解析交易[%s]的NFT数据失败: %v", txid, err)
					tapeJson = make(map[string]interface{})
				}
			}
		}

		// 从JSON中提取NFT文件信息
		nftFile, _ := tapeJson["file"].(string)
		if len(nftFile) == 72 {
			nftContractId = nftFile[:64]
			// 解析集合索引
			voutBytes := make([]byte, 4)
			for i := 0; i < 4; i++ {
				byteStr := nftFile[64+i*2 : 64+(i+1)*2]
				b, _ := strconv.ParseUint(byteStr, 16, 8)
				voutBytes[3-i] = byte(b)
			}
			collectionIndex = int(binary.BigEndian.Uint32(voutBytes))
		} else {
			nftContractId = txid
		}

		// 获取NFT基本信息
		nftInfo, err := logic.utxoSetDAO.GetNftUtxoByContractIdWithContext(ctx, nftContractId)
		if err != nil || nftInfo == nil {
			// 如果无法通过合约ID获取，尝试通过集合ID和索引获取
			nfts, err := logic.utxoSetDAO.GetNftsByCollectionAndIndex(ctx, nftContractId, collectionIndex)
			if err != nil || len(nfts) == 0 {
				log.WarnWithContextf(ctx, "获取NFT[%s]信息失败: %v", nftContractId, err)
				// 继续处理下一个交易
				continue
			}
			// 使用找到的第一个NFT信息
			nftInfo = nfts[0]
		}

		// 构建历史记录项目
		historyItem := nft.NftHistoryItem{
			Txid:               txid,
			CollectionId:       nftInfo.CollectionId,
			CollectionIndex:    nftInfo.CollectionIndex,
			CollectionName:     nftInfo.CollectionName,
			NftContractId:      nftInfo.NftContractId,
			NftName:            nftInfo.NftName,
			NftSymbol:          nftInfo.NftSymbol,
			NftDescription:     nftInfo.NftDescription,
			SenderAddresses:    senderAddresses,
			RecipientAddresses: recipientAddresses,
			TimeStamp:          timeStamp,
			UtcTime:            utcTime,
			NftIcon:            nftInfo.NftIcon,
		}

		historyItems = append(historyItems, historyItem)
	}

	// 构建响应数据
	response := &nft.NftHistoryResponse{
		Address:      address,
		ScriptHash:   nftScriptHash,
		HistoryCount: historyCount,
		Result:       historyItems,
	}

	log.InfoWithContextf(ctx, "成功获取地址[%s]的NFT历史记录，共%d条记录", address, historyCount)
	return response, nil
}
