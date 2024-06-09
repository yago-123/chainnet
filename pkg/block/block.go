package block

import (
	"time"
)

type Block struct {
	Timestamp     int64
	Transactions  []*Transaction
	PrevBlockHash []byte
	Hash          []byte
	Nonce         uint
}

func NewBlock(transactions []*Transaction, prevBlockHash []byte) *Block {
	block := &Block{
		time.Now().Unix(),
		transactions,
		prevBlockHash,
		[]byte{},
		0,
	}

	return block
}

func (block *Block) SetHashAndNonce(hash []byte, nonce uint) {
	block.Hash = hash[:]
	block.Nonce = nonce
}

func (block *Block) IsGenesisBlock() bool {
	return len(block.PrevBlockHash) == 0
}
