package mempool //nolint:testpackage // don't create separate package for tests

import (
	"github.com/stretchr/testify/require"
	"github.com/yago-123/chainnet/pkg/consensus/util"
	"github.com/yago-123/chainnet/tests/mocks/crypto/hash"
	"testing"

	"github.com/yago-123/chainnet/pkg/kernel"
	"github.com/yago-123/chainnet/pkg/script"

	"github.com/stretchr/testify/assert"
)

var tx1 = TxFeePair{ //nolint: gochecknoglobals // no need to lint this global variable
	Transaction: kernel.NewTransaction(
		[]kernel.TxInput{kernel.NewInput([]byte("id1"), 1, "sig", "pubkey1")},
		[]kernel.TxOutput{kernel.NewOutput(1, script.P2PK, "pubkey1")},
	),
	Fee: 10,
}

var tx2 = TxFeePair{ //nolint: gochecknoglobals // no need to lint this global variable
	Transaction: kernel.NewTransaction(
		[]kernel.TxInput{kernel.NewInput([]byte("id2"), 1, "sig", "pubkey2")},
		[]kernel.TxOutput{kernel.NewOutput(1, script.P2PK, "pubkey2")},
	),
	Fee: 2,
}

var tx3 = TxFeePair{ //nolint: gochecknoglobals // no need to lint this global variable
	Transaction: kernel.NewTransaction(
		[]kernel.TxInput{kernel.NewInput([]byte("id3"), 1, "sig", "pubkey3")},
		[]kernel.TxOutput{kernel.NewOutput(1, script.P2PK, "pubkey3")},
	),
	Fee: 3,
}

var tx4 = TxFeePair{ //nolint: gochecknoglobals // no need to lint this global variable
	Transaction: kernel.NewTransaction(
		[]kernel.TxInput{kernel.NewInput([]byte("id4"), 1, "sig", "pubkey4")},
		[]kernel.TxOutput{kernel.NewOutput(1, script.P2PK, "pubkey4")},
	),
	Fee: 1,
}

var tx5 = TxFeePair{ //nolint: gochecknoglobals // no need to lint this global variable
	Transaction: kernel.NewTransaction(
		[]kernel.TxInput{kernel.NewInput([]byte("id5"), 1, "sig", "pubkey5")},
		[]kernel.TxOutput{kernel.NewOutput(1, script.P2PK, "pubkey5")},
	),
	Fee: 9,
}

var tx6 = TxFeePair{ //nolint: gochecknoglobals // no need to lint this global variable
	Transaction: kernel.NewTransaction(
		[]kernel.TxInput{kernel.NewInput([]byte("id6"), 1, "sig", "pubkey6")},
		[]kernel.TxOutput{kernel.NewOutput(1, script.P2PK, "pubkey6")},
	),
	Fee: 6,
}

// transaction that share input with tx1
var txIncompatibleWithTx1 = TxFeePair{ //nolint: gochecknoglobals // no need to lint this global variable
	Transaction: kernel.NewTransaction(
		[]kernel.TxInput{kernel.NewInput([]byte("id1"), 1, "sig", "pubkey1")},
		[]kernel.TxOutput{kernel.NewOutput(1, script.P2PK, "pubkey9")},
	),
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
	txs, fee := mempool.RetrieveTransactions(1)
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
		txid, err := util.CalculateTxHash(v.Transaction, &hash.FakeHashing{})
		require.NoError(t, err)

		mempool.AppendTransaction(&kernel.Transaction{
			ID:   txid,
			Vin:  v.Transaction.Vin,
			Vout: v.Transaction.Vout,
		}, v.Fee)
	}

	txid, err := util.CalculateTxHash(txIncompatibleWithTx1.Transaction, &hash.FakeHashing{})
	require.NoError(t, err)

	// add tx that shares input with tx1
	mempool.AppendTransaction(&kernel.Transaction{
		ID:   txid,
		Vin:  txIncompatibleWithTx1.Transaction.Vin,
		Vout: txIncompatibleWithTx1.Transaction.Vout,
	}, txIncompatibleWithTx1.Fee)

	expectedInputSet := map[string][]string{
		"id1-1": []string{
			"Inputs:id11sigpubkey1Outputs:1\x005GBGp5uHwv OP_CHECKSIGpubkey1-hashed",
			"Inputs:id11sigpubkey1Outputs:1\x005GBGp5uHx4 OP_CHECKSIGpubkey9-hashed",
		},
		"id2-1": []string{"Inputs:id21sigpubkey2Outputs:1\x005GBGp5uHww OP_CHECKSIGpubkey2-hashed"},
		"id3-1": []string{"Inputs:id31sigpubkey3Outputs:1\x005GBGp5uHwx OP_CHECKSIGpubkey3-hashed"},
		"id4-1": []string{"Inputs:id41sigpubkey4Outputs:1\x005GBGp5uHwy OP_CHECKSIGpubkey4-hashed"},
		"id5-1": []string{"Inputs:id51sigpubkey5Outputs:1\x005GBGp5uHwz OP_CHECKSIGpubkey5-hashed"},
		"id6-1": []string{"Inputs:id61sigpubkey6Outputs:1\x005GBGp5uHx1 OP_CHECKSIGpubkey6-hashed"},
	}

	assert.Equal(t, expectedInputSet, mempool.inputSet)
}

func TestMemPoolOnBlockAddition(t *testing.T) {
	mempool := NewMemPool()

	for _, v := range txFeePairs {
		mempool.AppendTransaction(v.Transaction, v.Fee)
	}

	assert.Equal(t, 1, 2)
}
