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

func (bh *BlockHeader) IsEmpty() bool {
	return len(bh.Version) == 0 &&
		len(bh.PrevBlockHash) == 0 &&
		len(bh.MerkleRoot) == 0 &&
		bh.Height == 0 &&
		bh.Timestamp == 0 &&
		bh.Target == 0 &&
		bh.Nonce == 0
}

func (bh *BlockHeader) Assemble() []byte {
	data := [][]byte{
		[]byte(fmt.Sprintf("version %s", bh.Version)),
		[]byte(fmt.Sprintf("prev block hash %x", bh.PrevBlockHash)),
		[]byte(fmt.Sprintf("merkle root %x", bh.MerkleRoot)),
		[]byte(fmt.Sprintf("height %d", bh.Height)),
		[]byte(fmt.Sprintf("timestamp %d", bh.Timestamp)),
		[]byte(fmt.Sprintf("target %d", bh.Target)),
		[]byte(fmt.Sprintf("nonce %d", bh.Nonce)),
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
	return len(block.Header.PrevBlockHash) == 0 && block.Header.Height == 0
}
