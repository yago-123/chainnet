package mempool //nolint:testpackage // don't create separate package for tests

import (
	"github.com/yago-123/chainnet/pkg/kernel"
	"github.com/yago-123/chainnet/pkg/script"
	"testing"

	"github.com/stretchr/testify/assert"
)

var tx1 = TxFeePair{ //nolint: gochecknoglobals // no need to lint this global variable
	Transaction: kernel.NewTransaction(
		[]kernel.TxInput{kernel.NewInput([]byte("id1"), 1, "", "")},
		[]kernel.TxOutput{kernel.NewOutput(1, script.P2PK, "")},
	),
	Fee: 10,
}

var tx2 = TxFeePair{ //nolint: gochecknoglobals // no need to lint this global variable
	Transaction: kernel.NewTransaction(
		[]kernel.TxInput{kernel.NewInput([]byte("id2"), 1, "", "")},
		[]kernel.TxOutput{kernel.NewOutput(1, script.P2PK, "")},
	),
	Fee: 2,
}

var tx3 = TxFeePair{ //nolint: gochecknoglobals // no need to lint this global variable
	Transaction: kernel.NewTransaction(
		[]kernel.TxInput{kernel.NewInput([]byte("id3"), 1, "", "")},
		[]kernel.TxOutput{kernel.NewOutput(1, script.P2PK, "")},
	),
	Fee: 3,
}

var tx4 = TxFeePair{ //nolint: gochecknoglobals // no need to lint this global variable
	Transaction: kernel.NewTransaction(
		[]kernel.TxInput{kernel.NewInput([]byte("id4"), 1, "", "")},
		[]kernel.TxOutput{kernel.NewOutput(1, script.P2PK, "")},
	),
	Fee: 1,
}

var tx5 = TxFeePair{ //nolint: gochecknoglobals // no need to lint this global variable
	Transaction: kernel.NewTransaction(
		[]kernel.TxInput{kernel.NewInput([]byte("id5"), 1, "", "")},
		[]kernel.TxOutput{kernel.NewOutput(1, script.P2PK, "")},
	),
	Fee: 9,
}

var tx6 = TxFeePair{ //nolint: gochecknoglobals // no need to lint this global variable
	Transaction: kernel.NewTransaction(
		[]kernel.TxInput{kernel.NewInput([]byte("id6"), 1, "", "")},
		[]kernel.TxOutput{kernel.NewOutput(1, script.P2PK, "")},
	),
	Fee: 6,
}

var txFeePairs = []TxFeePair{tx1, tx2, tx3, tx4, tx5, tx6} //nolint: gochecknoglobals // no need to lint this global variable

func TestMemPool(t *testing.T) {
	mempool := NewMemPool()

	for _, v := range txFeePairs {
		mempool.AppendTransaction(v.Transaction, v.Fee)
	}
	assert.Equal(t, 6, mempool.Len())

	txs, fee := mempool.RetrieveTransactions(1)
	assert.Len(t, txs, 1)
	assert.Equal(t, 5, mempool.Len())
	assert.Equal(t, uint(10), fee)
	assert.Equal(t, []byte("id1"), txs[0].Vin[0].Txid)

	txs, fee = mempool.RetrieveTransactions(2)
	assert.Len(t, txs, 2)
	assert.Equal(t, 3, mempool.Len())
	assert.Equal(t, uint(15), fee)
	assert.Equal(t, []byte("id5"), txs[0].Vin[0].Txid)
	assert.Equal(t, []byte("id6"), txs[1].Vin[0].Txid)

	txs, fee = mempool.RetrieveTransactions(3)
	assert.Len(t, txs, 3)
	assert.Equal(t, 0, mempool.Len())
	assert.Equal(t, uint(6), fee)
	assert.Equal(t, []byte("id3"), txs[0].Vin[0].Txid)
	assert.Equal(t, []byte("id2"), txs[1].Vin[0].Txid)
	assert.Equal(t, []byte("id4"), txs[2].Vin[0].Txid)

	txs, fee = mempool.RetrieveTransactions(1)
	assert.Empty(t, txs)
	assert.Equal(t, 0, mempool.Len())
	assert.Equal(t, uint(0), fee)
}
