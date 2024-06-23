package block

import "chainnet/pkg/script"

// COINBASE_AMOUNT represents the reward for mining a block
const COINBASE_AMOUNT = 50

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

// TxInput represents the source of the transaction balance
type TxInput struct {
	// Txid is the transaction from which we are going to unlock the input balance
	Txid []byte

	// Vout is the index of the unspent transaction output (UTXO) that is going to be unlocked
	Vout uint

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
func NewInput(txid []byte, Vout uint, ScriptSig string, PubKey string) *TxInput {
	return &TxInput{
		Txid:      txid,
		Vout:      Vout,
		ScriptSig: ScriptSig,
		PubKey:    PubKey,
	}
}

func (in *TxInput) CanUnlockOutputWith(pubKey string) bool {
	return in.PubKey == pubKey
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
	// todo() come up with mechanism for halving COINBASE_AMOUNT
	return NewOutput(COINBASE_AMOUNT, scriptType, pubKey)
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

func (out *TxOutput) LockWith(scriptType script.ScriptType, pubKey []byte) {
	out.ScriptPubKey = script.NewScript(scriptType, pubKey)
}

func (tx *Transaction) IsCoinbase() bool {
	return len(tx.Vin) == 0
}
