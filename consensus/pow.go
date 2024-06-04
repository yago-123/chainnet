package consensus

import (
	"bytes"
	"crypto/sha256"
	"math/big"
)

type ProofOfWork struct {
	block  *Block
	target *big.Int

	cfg *Config
}

func NewProofOfWork(block *Block, cfg *Config) *ProofOfWork {
	target := big.NewInt(1)
	target.Lsh(target, uint(256-MINING_DIFFICULTY))

	return &ProofOfWork{block, target, cfg}
}

func (pow *ProofOfWork) assembleProofData(nonce int) []byte {
	data := [][]byte{
		pow.block.PrevBlockHash,
		pow.block.Data,
		[]byte(string(pow.block.Timestamp)),
		[]byte(string(MINING_DIFFICULTY)),
		[]byte(string(nonce)),
	}

	return bytes.Join(data, []byte{})
}

func (pow *ProofOfWork) Calculate() (int, []byte) {
	var hashInt big.Int
	var hash [32]byte
	nonce := 0

	pow.cfg.logger.Infof("Mining the block containing: %s", string(pow.block.Data))
	for nonce < MAX_NONCE {
		data := pow.assembleProofData(nonce)
		hash = sha256.Sum256(data)
		hashInt.SetBytes(hash[:])

		if hashInt.Cmp(pow.target) == -1 {
			break
		}

		nonce++
	}

	return nonce, hash[:]
}

func (pow *ProofOfWork) Validate() bool {
	var hashInt big.Int

	data := pow.assembleProofData(pow.block.Nonce)
	hash := sha256.Sum256(data)
	hashInt.SetBytes(hash[:])

	return hashInt.Cmp(pow.target) == -1
}
