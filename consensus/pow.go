package consensus

import (
	"bytes"
	"chainnet/block"
	"chainnet/config"
	"crypto/sha256"
	"math/big"
)

type ProofOfWork struct {
	target *big.Int

	cfg *config.Config
}

func NewProofOfWork(cfg *config.Config) *ProofOfWork {
	target := big.NewInt(1)
	target.Lsh(target, uint(256-cfg.DifficultyPoW))

	return &ProofOfWork{target, cfg}
}

func (pow *ProofOfWork) assembleProofData(block *block.Block, nonce uint) []byte {
	data := [][]byte{
		block.PrevBlockHash,
		block.Data,
		[]byte(string(block.Timestamp)),
		[]byte(string(pow.cfg.DifficultyPoW)),
		[]byte(string(nonce)),
	}

	return bytes.Join(data, []byte{})
}

func (pow *ProofOfWork) Calculate(block *block.Block) ([]byte, uint) {
	var hashInt big.Int
	var hash [32]byte
	nonce := uint(0)

	pow.cfg.Logger.Infof("Mining the block containing: %s", string(block.Data))
	for nonce < pow.cfg.MaxNoncePoW {
		data := pow.assembleProofData(block, nonce)
		hash = sha256.Sum256(data)
		hashInt.SetBytes(hash[:])

		if hashInt.Cmp(pow.target) == -1 {
			break
		}

		nonce++
	}

	return hash[:], nonce
}

func (pow *ProofOfWork) Validate(block *block.Block) bool {
	var hashInt big.Int

	data := pow.assembleProofData(block, block.Nonce)
	hash := sha256.Sum256(data)
	hashInt.SetBytes(hash[:])

	return hashInt.Cmp(pow.target) == -1
}
