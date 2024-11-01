package kernel

import (
	"bytes"
	"fmt"
	"math/rand/v2"

	"github.com/btcsuite/btcutil/base58"

	"github.com/yago-123/chainnet/pkg/script"
)

// SignatureType represents the different signatures that can be performed over a transaction
type SignatureType byte

const (
	// sign everything
	SighashAll SignatureType = iota
	// SIGHASH_NONE
	// SIGHASH_SINGLE
	// SIGHASH_ANYONECANPAY
)

// Transaction represents the atomic unit of the blockchain
type Transaction struct {
	// ID is the hash of the transaction
	ID []byte

	// Vin are the sources from which the transaction is going to be funded
	Vin []TxInput

	// Vout are the destination of the funds
	Vout []TxOutput

	// Version, lock time...
}

// NewTransaction creates a new transaction with the given inputs and outputs
func NewTransaction(inputs []TxInput, outputs []TxOutput) *Transaction {
	return &Transaction{ID: nil, Vin: inputs, Vout: outputs}
}

// NewCoinbaseTransaction creates a new transaction that pays the miners for their work
func NewCoinbaseTransaction(to string, reward, txFee uint) *Transaction {
	input := NewCoinbaseInput()
	rewardOutput := NewCoinbaseOutput(reward, script.P2PK, to)

	// if there is tx fee, make sure to add it to the rewardOutput
	if txFee > 0 {
		return NewTransaction([]TxInput{input}, []TxOutput{
			rewardOutput,
			NewCoinbaseOutput(txFee, script.P2PK, to),
		})
	}

	return NewTransaction([]TxInput{input}, []TxOutput{rewardOutput})
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

// HaveInputs checks if the transaction has any inputs
func (tx *Transaction) HaveInputs() bool {
	return len(tx.Vin) > 0
}

// HaveOutputs checks if the transaction has any outputs
func (tx *Transaction) HaveOutputs() bool {
	return len(tx.Vout) > 0
}

// todo() in theory, coinbase tx should also be at index 0
// IsCoinbase checks if the transaction is a coinbase transaction
func (tx *Transaction) IsCoinbase() bool {
	return len(tx.Vin) == 1 && len(tx.Vin[0].Txid) == 0
}

func (tx *Transaction) OutputAmount() uint {
	var amount uint
	for _, out := range tx.Vout {
		amount += out.Amount
	}

	return amount
}

func (tx *Transaction) String() string {
	inputs := ""
	for _, in := range tx.Vin {
		inputs += fmt.Sprintf("- %s\n", in.String())
	}

	outputs := ""
	for _, out := range tx.Vout {
		outputs += fmt.Sprintf("- %s\n", out.String())
	}

	msg := fmt.Sprintf("ID: %x\n", tx.ID)
	msg = fmt.Sprintf("%s%s", msg, inputs)
	msg = fmt.Sprintf("%s%s", msg, outputs)

	return msg
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

// UniqueTxoKey represents the equivalent of UniqueKey for UTXO but for the TxInput, which would be
// the STXO or Spent Transaction Output. This method is used in the UTXO set in order to remove those
// utxo that are being spent
func (in *TxInput) UniqueTxoKey() string {
	return fmt.Sprintf("%x-%d", in.Txid, in.Vout)
}

// NewCoinbaseInput creates a special transaction input called a Coinbase input. This type of input represents
// the source of new coins created during mining (i.e., it does not come from previous transactions). To avoid
// potential hash collisions (where identical transactions could produce the same hash), some randomness is
// introduced into the ScriptSig field
func NewCoinbaseInput() TxInput {
	return TxInput{
		// introduce randomness to prevent hash collisions
		ScriptSig: fmt.Sprintf("%d", rand.Int64()), //nolint:gosec // using math/rand in this case is OK
	}
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

// CanUnlockOutputWith checks if the input can unlock the output
func (in *TxInput) CanUnlockOutputWith(pubKey string) bool {
	return in.PubKey == pubKey
}

// UnlockWith solves the challenge presented by the output in order to unlock the funds
func (in *TxInput) UnlockWith(scriptSig string) {
	in.ScriptSig = scriptSig
}

// EqualInput checks if the input is the same as the given input
func (in *TxInput) EqualInput(input TxInput) bool {
	return bytes.Equal(in.Txid, input.Txid) && in.Vout == input.Vout
}

func (in *TxInput) String() string {
	return fmt.Sprintf(
		"TxInput: id %x-%d from %s, scriptSig: %s",
		in.Txid,
		in.Vout,
		base58.Encode([]byte(in.PubKey)),
		in.ScriptSig,
	)
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

// NewCoinbaseOutput creates a new output for the coinbase transaction
func NewCoinbaseOutput(amount uint, scriptType script.ScriptType, pubKey string) TxOutput {
	return NewOutput(amount, scriptType, pubKey)
}

// NewOutput creates a new output for the transaction
func NewOutput(amount uint, scriptType script.ScriptType, pubKey string) TxOutput {
	return TxOutput{
		Amount:       amount,
		ScriptPubKey: script.NewScript(scriptType, []byte(pubKey)),
		PubKey:       pubKey,
	}
}

// canBeUnlockedWith checks if the output can be unlocked with the given public key
func (out *TxOutput) CanBeUnlockedWith(pubKey string) bool {
	return out.PubKey == pubKey
}

func (out *TxOutput) String() string {
	return fmt.Sprintf(
		"TxOutput: %d to %s, unlocking script %s",
		out.Amount,
		base58.Encode([]byte(out.PubKey)),
		out.ScriptPubKey,
	)
}
