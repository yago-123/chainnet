package block

type Block struct {
	Timestamp     int64
	Transactions  []*Transaction
	PrevBlockHash []byte
	Target        uint
	Nonce         uint
	Hash          []byte
}

func NewBlock(transactions []*Transaction, prevBlockHash []byte) *Block {
	block := &Block{
		Transactions:  transactions,
		PrevBlockHash: prevBlockHash,
		Target:        0,
		Nonce:         0,
		Hash:          []byte{},
	}

	return block
}

func NewGenesisBlock(transactions []*Transaction) *Block {
	return NewBlock(transactions, []byte{})
}

func (block *Block) SetHashAndNonce(hash []byte, nonce uint) {
	block.Hash = hash
	block.Nonce = nonce
}

func (block *Block) IsGenesisBlock() bool {
	return len(block.PrevBlockHash) == 0
}

// todo() make coinbase transaction change amount based on block height
func NewCoinbaseTransaction(to, data string) *Transaction {
	txin := TxInput{Txid: []byte{}, Vout: -1, ScriptSig: data}
	txout := TxOutput{Amount: COINBASE_AMOUNT, ScriptPubKey: to}
	tx := Transaction{ID: nil, Vin: []TxInput{txin}, Vout: []TxOutput{txout}}

	return &tx
}

func NewTransaction(inputs []TxInput, outputs []TxOutput) *Transaction {
	return &Transaction{ID: nil, Vin: inputs, Vout: outputs}
}

func (tx *Transaction) SetID(hash []byte) {
	tx.ID = hash
}
