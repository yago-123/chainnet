package consensus

import (
	"bytes"
	"chainnet/pkg/block"
	"chainnet/pkg/crypto/hash"
	"errors"
	"math/big"
)

type ProofOfWork struct {
	target  uint
	hashing hash.Hashing
}

func NewProofOfWork(target uint, hashing hash.Hashing) *ProofOfWork {
	return &ProofOfWork{target: target, hashing: hashing}
}

func (pow *ProofOfWork) CalculateBlockHash(b *block.Block) ([]byte, uint, error) {
	var hashInt big.Int
	var hash []byte

	hashDifficulty := big.NewInt(1)
	hashDifficulty.Lsh(hashDifficulty, uint(256-b.Target))

	maxNonce := ^uint(0)
	nonce := uint(0)

	txsID := pow.hashTransactionIDs(b.Transactions)
	for nonce < maxNonce {
		data := b.Assemble(nonce, txsID)
		hash = pow.hashing.Hash(data)

		// todo() recheck this part
		if hashInt.Cmp(hashDifficulty) == -1 {
			return hash[:], nonce, nil
		}

		nonce++
	}

	return []byte{}, 0, errors.New("could not find hash for block")
}

func (pow *ProofOfWork) ValidateBlock(b *block.Block) bool {
	data := b.Assemble(b.Nonce, pow.hashTransactionIDs(b.Transactions))
	// todo() add more validations

	return pow.hashing.Verify(b.Hash, data)
}

func (pow *ProofOfWork) ValidateTx(tx *block.Transaction) bool {
	data := tx.Assemble()
	return pow.hashing.Verify(tx.ID, data)
}

func (pow *ProofOfWork) CalculateTxHash(tx *block.Transaction) ([]byte, error) {
	data := tx.Assemble()
	return pow.hashing.Hash(data), nil
}

func (pow *ProofOfWork) hashTransactionIDs(transactions []*block.Transaction) []byte {
	var txHashes [][]byte

	for _, tx := range transactions {
		txHashes = append(txHashes, tx.ID)
	}

	return pow.hashing.Hash(bytes.Join(txHashes, []byte{}))
}
