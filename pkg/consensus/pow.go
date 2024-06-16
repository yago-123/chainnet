package consensus

import (
	"bytes"
	"chainnet/pkg/block"
	"crypto/sha256"
	"errors"
	"fmt"
	"math/big"
	"reflect"
)

type ProofOfWork struct {
	target uint
}

func NewProofOfWork(target uint) *ProofOfWork {
	return &ProofOfWork{target: target}
}

func (pow *ProofOfWork) CalculateBlockHash(b *block.Block) ([]byte, uint, error) {
	var hashInt big.Int
	var hash [32]byte

	hashDifficulty := big.NewInt(1)
	hashDifficulty.Lsh(hashDifficulty, uint(256-b.Target))

	maxNonce := ^uint(0)
	nonce := uint(0)

	for nonce < maxNonce {
		data := pow.assembleProofBlock(b, nonce)
		hash = sha256.Sum256(data)

		if hashInt.Cmp(hashDifficulty) == -1 {
			return hash[:], nonce, nil
		}

		nonce++
	}

	return []byte{}, 0, errors.New("could not find hash for block")
}

func (pow *ProofOfWork) ValidateBlock(b *block.Block) bool {
	data := pow.assembleProofBlock(b, b.Nonce)
	hash := sha256.Sum256(data)

	return reflect.DeepEqual(hash, b.Hash)
}

func (pow *ProofOfWork) ValidateTx(tx *block.Transaction) bool {
	data := pow.assembleProofTx(tx)
	hash := sha256.Sum256(data)

	return reflect.DeepEqual(hash, tx.ID)
}

func (pow *ProofOfWork) CalculateTxHash(tx *block.Transaction) ([]byte, error) {
	data := pow.assembleProofTx(tx)
	hash := sha256.Sum256(data)

	return hash[:], nil
}

func (pow *ProofOfWork) assembleProofBlock(b *block.Block, nonce uint) []byte {
	data := [][]byte{
		b.PrevBlockHash,
		pow.hashTransactionIDs(b.Transactions),
		[]byte(fmt.Sprintf("%d", b.Target)),
		[]byte(fmt.Sprintf("%d", b.Timestamp)),
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
