package consensus

import (
	"bytes"
	"chainnet/pkg/block"
	"crypto/sha256"
	"errors"
	"fmt"
	"math/big"
	"reflect"
	"time"
)

type ProofOfWork struct {
	target uint
}

func NewProofOfWork(target uint) *ProofOfWork {
	return &ProofOfWork{target: target}
}

func (pow *ProofOfWork) CalculateBlockHash(b *block.Block) (*block.Block, error) {
	var hashInt big.Int
	var hash [32]byte

	hashDifficulty := big.NewInt(1)
	hashDifficulty.Lsh(hashDifficulty, uint(256-b.Target))

	b.Target = pow.target

	for {
		timestamp := time.Now().Unix()
		maxNonce := ^uint(0)
		nonce := uint(0)

		for nonce < maxNonce {
			b.Timestamp = timestamp
			b.Nonce = nonce

			data := pow.assembleProofBlock(b, nonce, timestamp)
			hash = sha256.Sum256(data)

			if hashInt.Cmp(hashDifficulty) == -1 {
				b.Timestamp = timestamp
				b.Nonce = nonce
				b.Hash = hash[:]

				return b, nil
			}

			nonce++
		}

	}

	return &block.Block{}, errors.New("could not find hash for block")
}

func (pow *ProofOfWork) ValidateBlock(b *block.Block) bool {
	data := pow.assembleProofBlock(b, b.Nonce, b.Timestamp)
	hash := sha256.Sum256(data)

	return reflect.DeepEqual(hash, b.Hash)
}

func (pow *ProofOfWork) ValidateTx(tx *block.Transaction) bool {
	data := pow.assembleProofTx(tx)
	hash := sha256.Sum256(data)

	return reflect.DeepEqual(hash, tx.ID)
}

func (pow *ProofOfWork) CalculateTxHash(tx *block.Transaction) (*block.Transaction, error) {
	data := pow.assembleProofTx(tx)
	hash := sha256.Sum256(data)

	tx.ID = hash[:]

	return tx, nil
}

func (pow *ProofOfWork) assembleProofBlock(block *block.Block, nonce uint, timestamp int64) []byte {
	data := [][]byte{
		block.PrevBlockHash,
		pow.hashTransactionIDs(block.Transactions),
		[]byte(fmt.Sprintf("%d", block.Target)),
		[]byte(fmt.Sprintf("%d", timestamp)),
		[]byte(fmt.Sprintf("%d", nonce)),
	}

	return bytes.Join(data, []byte{})
}

func (pow *ProofOfWork) assembleProofTx(tx *block.Transaction) []byte {
	var data []byte

	for _, input := range tx.Vin {
		data = append(data, input.Txid...)
		data = append(data, []byte(fmt.Sprintf("%d", input.Vout))...)
		data = append(data, []byte(input.ScriptSig)...)
	}

	for _, output := range tx.Vout {
		data = append(data, []byte(fmt.Sprintf("%d", output.Amount))...)
		data = append(data, []byte(output.ScriptPubKey)...)
	}

	return data
}

func (pow *ProofOfWork) hashTransactionIDs(transactions []*block.Transaction) []byte {
	var txHashes [][]byte
	var txHash [32]byte

	for _, tx := range transactions {
		txHashes = append(txHashes, tx.ID)
	}

	txHash = sha256.Sum256(bytes.Join(txHashes, []byte{}))

	return txHash[:]
}
