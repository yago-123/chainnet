package kernel

import (
	"bytes"
	"chainnet/pkg/script"
	"fmt"
)

// CoinbaseReward represents the reward for mining a kernel
const CoinbaseReward = 50

// SignatureType represents the different signatures that can be performed over a transaction
type SignatureType byte

const (
	// sign everything
	SighashAll SignatureType = iota
	// SIGHASH_NONE
	// SIGHASH_SINGLE
	// SIGHASH_ANYONECANPAY
)

// Transaction
type Transaction struct {
	// ID is the hash of the transaction
	ID []byte

	// Vin are the sources from which the transaction is going to be funded
	Vin []TxInput

	// Vout are the destination of the funds
	Vout []TxOutput

	// Version, lock time...
}

func NewTransaction(inputs []TxInput, outputs []TxOutput) *Transaction {
	return &Transaction{ID: nil, Vin: inputs, Vout: outputs}
}

func NewCoinbaseTransaction(to string) *Transaction {
	txin := NewCoinbaseInput()
	txout := NewCoinbaseOutput(script.P2PK, to)

	return NewTransaction([]TxInput{txin}, []TxOutput{txout})
}

func (tx *Transaction) SetID(hash []byte) {
	tx.ID = hash
}

// Assemble retrieves all the data from the transaction in order to perform operations
// like extracting the tx ID
func (tx *Transaction) Assemble() []byte {
	var data []byte

	if len(tx.Vin) > 0 {
		// add some static data to prevent hash collisions
		data = append(data, []byte("Inputs:")...)
	}

	for _, input := range tx.Vin {
		data = append(data, input.Txid...)
		data = append(data, []byte(fmt.Sprintf("%d", input.Vout))...)
		data = append(data, []byte(input.ScriptSig)...)
		data = append(data, []byte(input.PubKey)...)
	}

	if len(tx.Vout) > 0 {
		// add some static data to prevent hash collisions
		data = append(data, []byte("Outputs:")...)
	}

	for _, output := range tx.Vout {
		data = append(data, []byte(fmt.Sprintf("%d", output.Amount))...)
		data = append(data, []byte(output.ScriptPubKey)...)
		data = append(data, []byte(output.PubKey)...)
	}

	return data
}

// AssembleForSigning retrieves all data from the transaction to perform operations
// like generating the signature for unlocking outputs. Differs from Assemble because
// the input.ScriptSig field must be not included (otherwise the transaction can't be verified)
func (tx *Transaction) AssembleForSigning() []byte {
	var data []byte

	if len(tx.Vin) > 0 {
		// add some static data to prevent hash collisions
		data = append(data, []byte("Inputs:")...)
	}

	for _, input := range tx.Vin {
		data = append(data, input.Txid...)
		data = append(data, []byte(fmt.Sprintf("%d", input.Vout))...)
	}

	if len(tx.Vout) > 0 {
		// add some static data to prevent hash collisions
		data = append(data, []byte("Outputs:")...)
	}

	for _, output := range tx.Vout {
		data = append(data, []byte(fmt.Sprintf("%d", output.Amount))...)
		data = append(data, []byte(output.ScriptPubKey)...)
		data = append(data, []byte(output.PubKey)...)
	}

	return data
}

func (tx *Transaction) HaveInputs() bool {
	return len(tx.Vin) > 0
}

func (tx *Transaction) HaveOutputs() bool {
	return len(tx.Vout) > 0
}

// todo() in theory, coinbase tx should also be at index 0
func (tx *Transaction) IsCoinbase() bool {
	return len(tx.Vin) == 1 && len(tx.Vin[0].Txid) == 0
}

// TxInput represents the source of the transaction balance
type TxInput struct {
	// Txid is the transaction from which we are going to unlock the input balance
	Txid []byte

	// Vout is the index of the unspent transaction output (UTXO) that is going to be unlocked
	Vout uint

	// SignType is the type of signature required
	// SignType SignatureType

	// ScriptSig is the solved challenge presented by the output in order to unlock the funds
	ScriptSig string

	// PubKey is the public key that unlocked the ScriptSig
	// todo() eventually remove once we cleared ScriptSig
	PubKey string
}

// NewCoinbaseInput represents the source of the transactions for paying the miners (comes from nowhere)
func NewCoinbaseInput() TxInput {
	return TxInput{}
}

// NewInput represents the source of the transactions
func NewInput(txid []byte, vout uint, scriptSig string, pubKey string) TxInput {
	return TxInput{
		Txid:      txid,
		Vout:      vout,
		ScriptSig: scriptSig,
		PubKey:    pubKey,
	}
}

func (in *TxInput) CanUnlockOutputWith(pubKey string) bool {
	return in.PubKey == pubKey
}

func (in *TxInput) UnlockWith(scriptSig string) {
	in.ScriptSig = scriptSig
}

func (in *TxInput) EqualInput(input TxInput) bool {
	return bytes.Equal(in.Txid, input.Txid) && in.Vout == input.Vout
}

// TxOutput represents the destination of the transaction balance
type TxOutput struct {
	// Amount is the amount of funds that the output holds
	Amount uint

	// ScriptPubKey is the challenge that must be proved in order to unlock the output
	ScriptPubKey string

	// PubKey temporary field, must be extracted from ScriptPubKey directly
	// todo() use PubKeyHash eventually once the tests are migrated
	PubKey string
}

func NewCoinbaseOutput(scriptType script.ScriptType, pubKey string) TxOutput {
	// todo() come up with mechanism for halving CoinbaseReward
	return NewOutput(CoinbaseReward, scriptType, pubKey)
}

func NewOutput(amount uint, scriptType script.ScriptType, pubKey string) TxOutput {
	return TxOutput{
		Amount:       amount,
		ScriptPubKey: script.NewScript(scriptType, []byte(pubKey)),
		PubKey:       pubKey,
	}
}

func (out *TxOutput) CanBeUnlockedWith(pubKey string) bool {
	return out.PubKey == pubKey
}

// UnspentOutput
type UnspentOutput struct {
	TxID   []byte
	OutIdx uint
	Output TxOutput
}

func (utxo *UnspentOutput) EqualInput(input TxInput) bool {
	return bytes.Equal(utxo.TxID, input.Txid) && utxo.OutIdx == input.Vout
}
