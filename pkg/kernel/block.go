package kernel

import (
	"bytes"
	"fmt"
)

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

func NewGenesisBlock(transactions []*Transaction) *Block {
	return NewBlock(transactions, []byte{})
}

func (block *Block) SetHashAndNonce(hash []byte, nonce uint) {
	block.Hash = hash
	block.Nonce = nonce
}

func (block *Block) IsGenesisBlock() bool {
	return len(block.PrevBlockHash) == 0
}

func (block *Block) Assemble(nonce uint, txsID []byte) []byte {
	data := [][]byte{
		block.PrevBlockHash,
		txsID,
		[]byte(fmt.Sprintf("%d", block.Target)),
		[]byte(fmt.Sprintf("%d", block.Timestamp)),
		[]byte(fmt.Sprintf("%d", nonce)),
	}

	return bytes.Join(data, []byte{})
}
