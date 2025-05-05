package block

// BlockHeader 区块头信息
type BlockHeader struct {
	Hash              string  `json:"hash"`
	Confirmations     int64   `json:"confirmations"`
	Height            int64   `json:"height"`
	Version           int32   `json:"version"`
	VersionHex        string  `json:"versionHex"`
	MerkleRoot        string  `json:"merkleroot"`
	NumTx             int32   `json:"num_tx"`
	Time              int64   `json:"time"`
	MedianTime        int64   `json:"mediantime"`
	Nonce             uint32  `json:"nonce"`
	Bits              string  `json:"bits"`
	Difficulty        float64 `json:"difficulty"`
	ChainWork         string  `json:"chainwork"`
	PreviousBlockHash string  `json:"previousblockhash"`
	NextBlockHash     string  `json:"nextblockhash"`
}

// BlockDetail 区块详细信息
type BlockDetail struct {
	BlockHeader
	Tx   []string `json:"tx"`
	Size int32    `json:"size"`
}

// ValidateBlockHeight 验证区块高度参数
func ValidateBlockHeight(height int64) error {
	if height < 0 {
		return ErrInvalidBlockHeight
	}
	return nil
}

// ValidateBlockHash 验证区块哈希参数
func ValidateBlockHash(hash string) error {
	if len(hash) == 0 {
		return ErrEmptyBlockHash
	}
	return nil
}

// 错误定义
var (
	ErrInvalidBlockHeight = NewBlockError("区块高度必须大于等于0")
	ErrEmptyBlockHash     = NewBlockError("区块哈希不能为空")
)

// BlockError 区块错误
type BlockError struct {
	Message string
}

// NewBlockError 创建区块错误
func NewBlockError(message string) *BlockError {
	return &BlockError{Message: message}
}

func (e *BlockError) Error() string {
	return e.Message
}
