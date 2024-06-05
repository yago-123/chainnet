package block

import (
	"time"
)

type Block struct {
	Timestamp     int64
	Data          []byte
	PrevBlockHash []byte
	Hash          []byte
	Nonce         uint
}

func NewBlock(data string, prevBlockHash []byte) *Block {
	block := &Block{
		time.Now().Unix(),
		[]byte(data),
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
