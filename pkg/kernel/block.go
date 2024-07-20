package kernel

import (
	"bytes"
	"fmt"
)

type BlockHeader struct {
	Version       []byte
	PrevBlockHash []byte
	MerkleRoot    []byte
	// todo(): use timestamp to determine the difficulty, in a 2 weeks period, if the number of blocks was
	// todo(): created too quick, it means that the difficult must be increased
	Timestamp int64
	Target    uint
	Nonce     uint
}

type Block struct {
	Header       BlockHeader
	Transactions []*Transaction
	Hash         []byte
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
