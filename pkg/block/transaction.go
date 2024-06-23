package block

const COINBASE_AMOUNT = 50

type Transaction struct {
	// ID is the hash of the transaction
	ID []byte

	// Vin are the sources from which the transaction is going to be funded
	Vin []TxInput

	// Vout are the destination of the funds
	Vout []TxOutput

	// Version, lock time
}

// TxInput represents the source of the transaction balance
type TxInput struct {
	// Txid is the transaction from which we are going to unlock the input balance
	Txid []byte

	// Vout is the index of the unspent transaction output (UTXO) that is going to be unlocked
	Vout int

	// ScriptSig is the solved challenge presented by the output in order to unlock the funds
	ScriptSig string

	// PubKey is the public key that unlocked the ScriptSig
	// todo() eventually remove once we cleared ScriptSig
	PubKey string
}

func (in *TxInput) CanUnlockOutputWith(pubKey string) bool {
	return in.PubKey == pubKey
}

// TxOutput represents the destination of the transaction balance
type TxOutput struct {
	// Amount is the amount of funds that the output holds
	Amount int

	// ScriptPubKey is the challenge that must be proved in order to unlock the output
	ScriptPubKey string

	// PubKey temporary field, must be extracted from ScriptPubKey directly
	// todo() use PubKeyHash eventually once the tests are migrated
	PubKey string
}

func (out *TxOutput) CanBeUnlockedWith(pubKey string) bool {
	return out.PubKey == pubKey
}

func (tx *Transaction) IsCoinbase() bool {
	return len(tx.Vin) == 0
}
