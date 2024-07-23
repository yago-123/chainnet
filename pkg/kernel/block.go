package kernel

import (
	"bytes"
	"fmt"
)

const MaxNumberTxsPerBlock = 16000

type BlockHeader struct {
	Version       []byte
	PrevBlockHash []byte
	MerkleRoot    []byte
	Height        uint
	// todo(): use timestamp to determine the difficulty, in a 2 weeks period, if the number of blocks was
	// todo(): created too quick, it means that the difficult must be increased
	Timestamp int64
	Target    uint
	Nonce     uint
}

func NewBlockHeader(version []byte, timestamp int64, merkleRoot []byte, height uint, prevBlockHash []byte, target, nonce uint) *BlockHeader {
	return &BlockHeader{
		Version:       version,
		Timestamp:     timestamp,
		MerkleRoot:    merkleRoot,
		Height:        height,
		PrevBlockHash: prevBlockHash,
		Target:        target,
		Nonce:         nonce,
	}
}

func (bh *BlockHeader) SetNonce(nonce uint) {
	bh.Nonce = nonce
}

func (bh *BlockHeader) SetTimestamp(timestamp int64) {
	bh.Timestamp = timestamp

}

func (bh *BlockHeader) Assemble() []byte {
	data := [][]byte{
		[]byte(fmt.Sprintf("version %s", bh.Version)),
		[]byte(fmt.Sprintf("prev block hash %s", bh.PrevBlockHash)),
		[]byte(fmt.Sprintf("merkle root %s", bh.MerkleRoot)),
		[]byte(fmt.Sprintf("height %d", bh.Height)),
		[]byte(fmt.Sprintf("timestamp %d", bh.Timestamp)),
		[]byte(fmt.Sprintf("target %d", bh.Target)),
		[]byte(fmt.Sprintf("nonce %d", bh.Nonce)),
	}

	return bytes.Join(data, []byte{})
}

func (bh *BlockHeader) AssembleWithNonce(nonce uint) []byte {
	data := [][]byte{
		[]byte(fmt.Sprintf("version %s", bh.Version)),
		[]byte(fmt.Sprintf("prev block hash %s", bh.PrevBlockHash)),
		[]byte(fmt.Sprintf("merkle root %s", bh.MerkleRoot)),
		[]byte(fmt.Sprintf("height %d", bh.Height)),
		[]byte(fmt.Sprintf("timestamp %d", bh.Timestamp)),
		[]byte(fmt.Sprintf("target %d", bh.Target)),
		[]byte(fmt.Sprintf("nonce %d", nonce)),
	}

	return bytes.Join(data, []byte{})
}

type Block struct {
	Header       *BlockHeader
	Transactions []*Transaction
	Hash         []byte
}

func NewBlock(blockHeader *BlockHeader, transactions []*Transaction, blockHash []byte) *Block {
	block := &Block{
		Header:       blockHeader,
		Transactions: transactions,
		Hash:         blockHash,
	}

	return block
}

func NewGenesisBlock(blockHeader *BlockHeader, transactions []*Transaction, blockHash []byte) *Block {
	return NewBlock(blockHeader, transactions, blockHash)
}

func (block *Block) IsGenesisBlock() bool {
	return len(block.Header.PrevBlockHash) == 0
}
