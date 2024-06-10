package consensus

import (
	"bytes"
	"chainnet/config"
	"chainnet/pkg/block"
	"crypto/sha256"
	"fmt"
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
		pow.hashTransactions(block.Transactions),
		[]byte(fmt.Sprintf("%d", block.Timestamp)),
		[]byte(fmt.Sprintf("%d", pow.cfg.DifficultyPoW)),
		[]byte(fmt.Sprintf("%d", nonce)),
	}

	return bytes.Join(data, []byte{})
}

func (pow *ProofOfWork) hashTransactions(transactions []*block.Transaction) []byte {
	var txHashes [][]byte
	var txHash [32]byte

	for _, tx := range transactions {
		txHashes = append(txHashes, tx.ID)
	}

	txHash = sha256.Sum256(bytes.Join(txHashes, []byte{}))

	return txHash[:]
}

func (pow *ProofOfWork) Calculate(block *block.Block) ([]byte, uint) {
	var hashInt big.Int
	var hash [32]byte
	nonce := uint(0)

	pow.cfg.Logger.Infof("Mining the block containing %d transactions", len(block.Transactions))
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
