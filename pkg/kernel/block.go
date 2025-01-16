package kernel

import (
	"bytes"
	"fmt"
)

const (
	// ChainnetCoinAmount number of smaller units (Channoshis) that represent 1 Chainnet coin
	ChainnetCoinAmount   = 100000000
	MaxNumberTxsPerBlock = 300
)

type BlockHeader struct {
	Version       []byte
	PrevBlockHash []byte
	MerkleRoot    []byte
	Height        uint
	// todo(): use timestamp to determine the difficulty, in a 2 weeks period, if the number of blocks was
	// todo(): created too quick, it means that the difficult must be increased
	Timestamp int64
	// todo(): target could be removed now, mining difficulty can already be determined dynamically
	Target uint
	Nonce  uint
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

func (bh *BlockHeader) IsGenesisHeader() bool {
	return len(bh.PrevBlockHash) == 0 && bh.Height == 0
}

func (bh *BlockHeader) String() string {
	return fmt.Sprintf("header(version: %s, prev block hash: %x, merkle root: %x, height: %d, timestamp: %d, target: %d, nonce: %d)",
		bh.Version, bh.PrevBlockHash, bh.MerkleRoot, bh.Height, bh.Timestamp, bh.Target, bh.Nonce)
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

func ConvertFromChannoshisToCoins(channoshis uint) float64 {
	if channoshis == 0 {
		return 0.0
	}

	return float64(channoshis) / float64(ChainnetCoinAmount)
}

func ConvertFromCoinsToChannoshis(coins float64) uint {
	return uint(coins * float64(ChainnetCoinAmount))
}
