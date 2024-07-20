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

func NewBlockHeader(version []byte, timestamp uint, merkleRoot []byte, prevBlockHash []byte, target, nonce uint) *BlockHeader {
	return &BlockHeader{
		Version:       []byte("1"),
		Timestamp:     0,
		MerkleRoot:    []byte{},
		PrevBlockHash: prevBlockHash,
		Target:        0,
		Nonce:         0,
	}
}

type Block struct {
	Header       *BlockHeader
	Transactions []*Transaction
	Hash         []byte
}

func NewBlock(blockHeader *BlockHeader, transactions []*Transaction) *Block {
	block := &Block{
		Header:       blockHeader,
		Transactions: transactions,
		Hash:         []byte{},
	}

	return block
}

func NewGenesisBlock(blockHeader *BlockHeader, transactions []*Transaction) *Block {
	return NewBlock(blockHeader, transactions)
}

func (block *Block) SetHashAndNonce(hash []byte, nonce uint) {
	block.Hash = hash
	block.Header.Nonce = nonce
}

func (block *Block) IsGenesisBlock() bool {
	return len(block.Header.PrevBlockHash) == 0
}

func (block *Block) Assemble(nonce uint, txsID []byte) []byte {
	data := [][]byte{
		block.Header.PrevBlockHash,
		txsID,
		[]byte(fmt.Sprintf("%d", block.Header.Target)),
		[]byte(fmt.Sprintf("%d", block.Header.Timestamp)),
		[]byte(fmt.Sprintf("%d", nonce)),
	}

	return bytes.Join(data, []byte{})
}
