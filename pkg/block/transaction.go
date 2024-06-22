package block

const COINBASE_AMOUNT = 50

type Transaction struct {
	ID   []byte
	Vin  []TxInput
	Vout []TxOutput
}

type TxInput struct {
	Txid      []byte
	Vout      int
	ScriptSig string
}

func (in *TxInput) CanUnlockOutputWith(address string) bool {
	return in.ScriptSig == address
}

type TxOutput struct {
	Amount       int
	ScriptPubKey string
}

func (out *TxOutput) CanBeUnlockedWith(address string) bool {
	return out.ScriptPubKey == address
}

func (tx *Transaction) IsCoinbase() bool {
	return len(tx.Vin) == 0
}
