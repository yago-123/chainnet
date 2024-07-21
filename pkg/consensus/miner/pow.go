package miner

import (
	"chainnet/pkg/crypto/hash"
	"chainnet/pkg/kernel"
	"context"
	"errors"
	"fmt"
	"math/big"
)

const HashLength = 256

type ProofOfWork struct {
	target  uint
	hashing hash.Hashing
}

func NewProofOfWork(target uint, hashing hash.Hashing) *ProofOfWork {
	return &ProofOfWork{target: target, hashing: hashing}
}

func (pow *ProofOfWork) CalculateBlockHash(bh *kernel.BlockHeader, ctx context.Context) ([]byte, uint, error) {
	var err error
	var hash []byte
	var hashInt big.Int

	hashDifficulty := big.NewInt(1)
	hashDifficulty.Lsh(hashDifficulty, HashLength-bh.Target)

	maxNonce := ^uint(0)
	nonce := uint(0)

	for nonce < maxNonce {
		data := bh.Assemble()
		hash, err = pow.hashing.Hash(data)
		if err != nil {
			return []byte{}, 0, fmt.Errorf("could not hash block: %w", err)
		}

		// todo() recheck this part
		if hashInt.Cmp(hashDifficulty) == -1 {
			return hash, nonce, nil
		}

		nonce++
	}

	return []byte{}, 0, errors.New("could not find hash for kernel")
}

func (pow *ProofOfWork) CalculateTxHash(tx *kernel.Transaction) ([]byte, error) {
	data := tx.Assemble()
	return pow.hashing.Hash(data)
}
