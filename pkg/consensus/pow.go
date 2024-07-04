package consensus

import (
	"bytes"
	"chainnet/pkg/crypto/hash"
	"chainnet/pkg/kernel"
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

func (pow *ProofOfWork) CalculateBlockHash(b *kernel.Block) ([]byte, uint, error) {
	var hashInt big.Int
	var hash []byte

	hashDifficulty := big.NewInt(1)
	hashDifficulty.Lsh(hashDifficulty, HashLength-b.Target)

	maxNonce := ^uint(0)
	nonce := uint(0)

	txsID, err := pow.hashTransactionIDs(b.Transactions)
	if err != nil {
		return []byte{}, 0, fmt.Errorf("could not hash transactions: %w", err)
	}

	for nonce < maxNonce {
		data := b.Assemble(nonce, txsID)
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

func (pow *ProofOfWork) hashTransactionIDs(transactions []*kernel.Transaction) ([]byte, error) {
	var txHashes [][]byte

	for _, tx := range transactions {
		txHashes = append(txHashes, tx.ID)
	}

	h, err := pow.hashing.Hash(bytes.Join(txHashes, []byte{}))
	if err != nil {
		return []byte{}, err
	}

	return h, nil
}
