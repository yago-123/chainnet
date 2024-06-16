package block

type Block struct {
	Timestamp     int64
	Transactions  []*Transaction
	PrevBlockHash []byte
	Target        uint
	Nonce         uint
	Hash          []byte
}

func NewBlock(transactions []*Transaction, prevBlockHash []byte) *Block {
	block := &Block{
		Transactions:  transactions,
		PrevBlockHash: prevBlockHash,
		Target:        0,
		Nonce:         0,
		Hash:          []byte{},
	}

	return block
}

func (block *Block) SetHashAndNonce(hash []byte, nonce uint) {
	block.Hash = hash
	block.Nonce = nonce
}

func (block *Block) IsGenesisBlock() bool {
	return len(block.PrevBlockHash) == 0
}
