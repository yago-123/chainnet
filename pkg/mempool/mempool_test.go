package mempool //nolint:testpackage // don't create separate package for tests

import (
	"testing"

	"github.com/yago-123/chainnet/pkg/kernel"
	"github.com/yago-123/chainnet/pkg/script"

	"github.com/stretchr/testify/assert"
)

var tx1 = TxFeePair{ //nolint: gochecknoglobals // no need to lint this global variable
	Transaction: &kernel.Transaction{
		ID:   []byte("tx1"),
		Vin:  []kernel.TxInput{kernel.NewInput([]byte("id1"), 1, "sig", "pubkey1")},
		Vout: []kernel.TxOutput{kernel.NewOutput(1, script.P2PK, "pubkey1")},
	},
	Fee: 10,
}

var tx2 = TxFeePair{ //nolint: gochecknoglobals // no need to lint this global variable
	Transaction: &kernel.Transaction{
		ID:   []byte("tx2"),
		Vin:  []kernel.TxInput{kernel.NewInput([]byte("id2"), 1, "sig", "pubkey2")},
		Vout: []kernel.TxOutput{kernel.NewOutput(1, script.P2PK, "pubkey2")},
	},
	Fee: 2,
}

var tx3 = TxFeePair{ //nolint: gochecknoglobals // no need to lint this global variable
	Transaction: &kernel.Transaction{
		ID:   []byte("tx3"),
		Vin:  []kernel.TxInput{kernel.NewInput([]byte("id3"), 1, "sig", "pubkey3")},
		Vout: []kernel.TxOutput{kernel.NewOutput(1, script.P2PK, "pubkey3")},
	},
	Fee: 3,
}

var tx4 = TxFeePair{ //nolint: gochecknoglobals // no need to lint this global variable
	Transaction: &kernel.Transaction{
		ID:   []byte("tx4"),
		Vin:  []kernel.TxInput{kernel.NewInput([]byte("id4"), 1, "sig", "pubkey4")},
		Vout: []kernel.TxOutput{kernel.NewOutput(1, script.P2PK, "pubkey4")},
	},
	Fee: 1,
}

var tx5 = TxFeePair{ //nolint: gochecknoglobals // no need to lint this global variable
	Transaction: &kernel.Transaction{
		ID:   []byte("tx5"),
		Vin:  []kernel.TxInput{kernel.NewInput([]byte("id5"), 1, "sig", "pubkey5")},
		Vout: []kernel.TxOutput{kernel.NewOutput(1, script.P2PK, "pubkey5")},
	},
	Fee: 9,
}

var tx6 = TxFeePair{ //nolint: gochecknoglobals // no need to lint this global variable
	Transaction: &kernel.Transaction{
		ID:   []byte("tx6"),
		Vin:  []kernel.TxInput{kernel.NewInput([]byte("id6"), 1, "sig", "pubkey6")},
		Vout: []kernel.TxOutput{kernel.NewOutput(1, script.P2PK, "pubkey6")},
	},
	Fee: 6,
}

// transaction that share input with tx1
var txIncompatibleWithTx1 = TxFeePair{ //nolint: gochecknoglobals // no need to lint this global variable
	Transaction: &kernel.Transaction{
		ID:   []byte("txIncompatibleWithTx1"),
		Vin:  []kernel.TxInput{kernel.NewInput([]byte("id1"), 1, "sig", "pubkey1")},
		Vout: []kernel.TxOutput{kernel.NewOutput(1, script.P2PK, "pubkey9")},
	},
	Fee: 9,
}

var txFeePairs = []TxFeePair{tx1, tx2, tx3, tx4, tx5, tx6} //nolint: gochecknoglobals // no need to lint this global variable

func TestRetrieveTxsWithoutIncompatibilities(t *testing.T) {
	mempool := NewMemPool()
	// add 6 txs to the mempool
	for _, v := range txFeePairs {
		mempool.AppendTransaction(v.Transaction, v.Fee)
	}

	assert.Equal(t, 6, mempool.Len())

	// checks for RetrieveTransactions
	txs, fee := mempool.RetrieveTransactions(0)
	assert.Len(t, txs, 0)

	txs, fee = mempool.RetrieveTransactions(1)
	assert.Len(t, txs, 1)
	assert.Equal(t, uint(10), fee)
	assert.Equal(t, []byte("id1"), txs[0].Vin[0].Txid)

	txs, fee = mempool.RetrieveTransactions(2)
	assert.Len(t, txs, 2)
	assert.Equal(t, uint(19), fee)
	assert.Equal(t, []byte("id1"), txs[0].Vin[0].Txid)
	assert.Equal(t, []byte("id5"), txs[1].Vin[0].Txid)

	txs, fee = mempool.RetrieveTransactions(3)
	assert.Len(t, txs, 3)
	assert.Equal(t, uint(25), fee)
	assert.Equal(t, []byte("id1"), txs[0].Vin[0].Txid)
	assert.Equal(t, []byte("id5"), txs[1].Vin[0].Txid)
	assert.Equal(t, []byte("id6"), txs[2].Vin[0].Txid)

	txs, fee = mempool.RetrieveTransactions(1)
	assert.Len(t, txs, 1)
	assert.Equal(t, uint(10), fee)
	assert.Equal(t, []byte("id1"), txs[0].Vin[0].Txid)

	txs, fee = mempool.RetrieveTransactions(10)
	assert.Len(t, txs, 6)
}

func TestRetrieveTxsWithIncompatibilities(t *testing.T) {
	mempool := NewMemPool()

	for _, v := range txFeePairs {
		mempool.AppendTransaction(v.Transaction, v.Fee)
	}

	mempool.AppendTransaction(txIncompatibleWithTx1.Transaction, txIncompatibleWithTx1.Fee)

	txs, fee := mempool.RetrieveTransactions(3)
	assert.Len(t, txs, 3)
	assert.Equal(t, uint(25), fee)
	assert.Equal(t, []byte("id1"), txs[0].Vin[0].Txid)
	assert.Equal(t, []byte("id5"), txs[1].Vin[0].Txid)
	assert.Equal(t, []byte("id6"), txs[2].Vin[0].Txid)
}

func TestMemPoolInputSet(t *testing.T) {
	mempool := NewMemPool()

	// add 6 txs to the mempool
	for _, v := range txFeePairs {
		mempool.AppendTransaction(v.Transaction, v.Fee)
	}

	// add tx that shares input with tx1
	mempool.AppendTransaction(txIncompatibleWithTx1.Transaction, txIncompatibleWithTx1.Fee)

	expectedInputSet := map[string][]string{
		"id1-1": []string{
			"tx1",
			"txIncompatibleWithTx1",
		},
		"id2-1": []string{"tx2"},
		"id3-1": []string{"tx3"},
		"id4-1": []string{"tx4"},
		"id5-1": []string{"tx5"},
		"id6-1": []string{"tx6"},
	}

	assert.Equal(t, expectedInputSet, mempool.inputSet)
}

func TestMemPoolOnBlockAddition(t *testing.T) {
	mempool := NewMemPool()

	for _, v := range txFeePairs {
		mempool.AppendTransaction(v.Transaction, v.Fee)
	}

	mempool.AppendTransaction(txIncompatibleWithTx1.Transaction, txIncompatibleWithTx1.Fee)

	mempool.OnBlockAddition(
		&kernel.Block{
			Transactions: []*kernel.Transaction{
				tx1.Transaction,
				tx2.Transaction,
				tx6.Transaction,
			},
		},
	)

	expectedInputSet := map[string][]string{
		"id3-1": []string{"tx3"},
		"id4-1": []string{"tx4"},
		"id5-1": []string{"tx5"},
	}

	assert.Equal(t, expectedInputSet, mempool.inputSet)
}

func TestMemPoolCoinbaseTx(t *testing.T) {
	_ = NewMemPool()

	assert.Equal(t, 1, 2)
}

func TestMemPoolGenesisBlock(t *testing.T) {
	assert.Equal(t, 1, 2)
}
