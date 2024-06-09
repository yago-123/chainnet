package block

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

func (in *TxInput) CanUnlockOutputWith(unlockingData string) bool {
	return in.ScriptSig == unlockingData
}

type TxOutput struct {
	Amount       int
	ScriptPubKey string
}

func (out *TxOutput) CanBeUnlockedWith(unlockingData string) bool {
	return out.ScriptPubKey == unlockingData
}

func NewCoinbaseTx(to, data string) *Transaction {
	if data == "" {
		data = "Reward to " + to
	}

	txin := TxInput{[]byte{}, -1, data}
	txout := TxOutput{100, to}
	tx := Transaction{nil, []TxInput{txin}, []TxOutput{txout}}
	//tx.SetID()

	return &tx
}

func (tx *Transaction) IsCoinbase() bool {
	if len(tx.Vin) == 0 {
		return true
	}

	return false
}
