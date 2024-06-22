package block

import (
	"bytes"
	"chainnet/pkg/crypto/hash"
	"github.com/btcsuite/btcutil/base58"
)

const COINBASE_AMOUNT = 50

type Transaction struct {
	ID   []byte
	Vin  []TxInput
	Vout []TxOutput
}

type TxInput struct {
	Txid      []byte
	Vout      int
	Signature []byte
	PubKey    []byte
}

func (in *TxInput) UsesKey(pubKeyHash []byte, hashing hash.Hashing) bool {
	lockingHash := hashing.Hash(in.PubKey)

	return bytes.Equal(lockingHash, pubKeyHash)
}

type TxOutput struct {
	Amount     int
	PubKeyHash []byte
}

func (out *TxOutput) Lock(address []byte) {
	pubKeyHash := base58.Decode(string(address))
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-4]
	out.PubKeyHash = pubKeyHash
}

func (out *TxOutput) IsLockedWithKey(pubKeyHash []byte) bool {
	return bytes.Equal(out.PubKeyHash, pubKeyHash)
}

func (tx *Transaction) IsCoinbase() bool {
	return len(tx.Vin) == 0
}
