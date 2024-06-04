package block

import "time"

type Block struct {
	Timestamp     int64
	Data          []byte
	PrevBlockHash []byte
	Hash          []byte
	Nonce         int
}

func NewBlock(data string, prevBlockHash []byte, cfg *Config) *Block {
	block := &Block{
		time.Now().Unix(),
		[]byte(data),
		prevBlockHash,
		[]byte{},
		0,
	}

	pow := NewProofOfWork(block, cfg)
	nonce, hash := pow.Calculate()

	block.Hash = hash[:]
	block.Nonce = nonce

	return block
}
