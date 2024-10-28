package blockchain //nolint:testpackage // don't create separate package for tests

import (
	"github.com/yago-123/chainnet/config"
	"testing"

	"github.com/yago-123/chainnet/pkg/kernel"
	"github.com/yago-123/chainnet/pkg/script"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var b1 = &kernel.Block{ //nolint:gochecknoglobals // ignore linter in this case
	Transactions: []*kernel.Transaction{
		{
			ID: []byte("coinbase-transaction-block-1"),
			Vin: []kernel.TxInput{
				kernel.NewCoinbaseInput(),
			},
			Vout: []kernel.TxOutput{
				kernel.NewCoinbaseOutput(50, script.P2PK, "alice"), // <- spent by transaction-1-block-2
			},
		},
	},
}

var b2 = &kernel.Block{ //nolint:gochecknoglobals // ignore linter in this case
	Transactions: []*kernel.Transaction{
		{
			ID: []byte("coinbase-transaction-block-2"),
			Vin: []kernel.TxInput{
				kernel.NewCoinbaseInput(),
			},
			Vout: []kernel.TxOutput{
				kernel.NewCoinbaseOutput(50, script.P2PK, "bob"), // <- unspent
			},
		},
		{
			ID: []byte("transaction-1-block-2"),
			Vin: []kernel.TxInput{
				{
					Txid:      []byte("coinbase-transaction-block-1"),
					Vout:      0,
					ScriptSig: "",
					PubKey:    "",
				},
			},
			Vout: []kernel.TxOutput{
				{
					Amount:       25,
					ScriptPubKey: "",
					PubKey:       "mike", // <- spent by transaction-1-block-3
				},
				{
					Amount:       25,
					ScriptPubKey: "",
					PubKey:       "chris", // <- unspent
				},
			},
		},
	},
}

var b3 = &kernel.Block{ //nolint:gochecknoglobals // ignore linter in this case
	Transactions: []*kernel.Transaction{
		{
			ID: []byte("coinbase-transaction-block-3"),
			Vin: []kernel.TxInput{
				kernel.NewCoinbaseInput(),
			},
			Vout: []kernel.TxOutput{
				kernel.NewCoinbaseOutput(50, script.P2PK, "chris"), // <- unspent
			},
		},
		{
			ID: []byte("transaction-1-block-3"),
			Vin: []kernel.TxInput{
				{
					Txid:      []byte("transaction-1-block-2"),
					Vout:      0,
					ScriptSig: "",
					PubKey:    "",
				},
			},
			Vout: []kernel.TxOutput{
				{
					Amount:       25,
					ScriptPubKey: "",
					PubKey:       "dave", // <- unspent
				},
			},
		},
	},
}

func TestUTXOSet_AddBlock(t *testing.T) {
	utxos := NewUTXOSet(config.NewConfig())

	require.NoError(t, utxos.AddBlock(b1))
	require.NoError(t, utxos.AddBlock(b2))
	require.NoError(t, utxos.AddBlock(b3))

	require.Len(t, utxos.utxos, 4)

	val, ok := utxos.utxos["coinbase-transaction-block-2-0"]
	require.True(t, ok)
	assert.Equal(t, uint(0), val.OutIdx)
	assert.Equal(t, "bob", val.Output.PubKey)
	assert.Equal(t, uint(50), val.Output.Amount)

	val, ok = utxos.utxos["transaction-1-block-2-1"]
	require.True(t, ok)
	assert.Equal(t, uint(1), val.OutIdx)
	assert.Equal(t, "chris", val.Output.PubKey)
	assert.Equal(t, uint(25), val.Output.Amount)

	val, ok = utxos.utxos["coinbase-transaction-block-3-0"]
	require.True(t, ok)
	assert.Equal(t, uint(0), val.OutIdx)
	assert.Equal(t, "chris", val.Output.PubKey)
	assert.Equal(t, uint(50), val.Output.Amount)

	val, ok = utxos.utxos["transaction-1-block-3-0"]
	require.True(t, ok)
	assert.Equal(t, uint(0), val.OutIdx)
	assert.Equal(t, "dave", val.Output.PubKey)
	assert.Equal(t, uint(25), val.Output.Amount)
}

// same as TestUTXOSet_AddBlock but with OnBlockAddition method
func TestUTXOSet_OnBlockAddition(t *testing.T) {
	utxos := NewUTXOSet(config.NewConfig())

	utxos.OnBlockAddition(b1)
	utxos.OnBlockAddition(b2)
	utxos.OnBlockAddition(b3)

	require.Len(t, utxos.utxos, 4)

	val, ok := utxos.utxos["coinbase-transaction-block-2-0"]
	require.True(t, ok)
	assert.Equal(t, uint(0), val.OutIdx)
	assert.Equal(t, "bob", val.Output.PubKey)
	assert.Equal(t, uint(50), val.Output.Amount)

	val, ok = utxos.utxos["transaction-1-block-2-1"]
	require.True(t, ok)
	assert.Equal(t, uint(1), val.OutIdx)
	assert.Equal(t, "chris", val.Output.PubKey)
	assert.Equal(t, uint(25), val.Output.Amount)

	val, ok = utxos.utxos["coinbase-transaction-block-3-0"]
	require.True(t, ok)
	assert.Equal(t, uint(0), val.OutIdx)
	assert.Equal(t, "chris", val.Output.PubKey)
	assert.Equal(t, uint(50), val.Output.Amount)

	val, ok = utxos.utxos["transaction-1-block-3-0"]
	require.True(t, ok)
	assert.Equal(t, uint(0), val.OutIdx)
	assert.Equal(t, "dave", val.Output.PubKey)
	assert.Equal(t, uint(25), val.Output.Amount)
}

func TestUTXOSet_RetrieveInputsBalance(t *testing.T) {
	utxos := NewUTXOSet(config.NewConfig())

	utxos.OnBlockAddition(b1)
	utxos.OnBlockAddition(b2)
	utxos.OnBlockAddition(b3)

	balance, err := utxos.RetrieveInputsBalance([]kernel.TxInput{
		{Txid: []byte("coinbase-transaction-block-2"), Vout: 0},
	})
	require.NoError(t, err)
	require.Equal(t, uint(50), balance)

	balance, err = utxos.RetrieveInputsBalance([]kernel.TxInput{
		{Txid: []byte("coinbase-transaction-block-2"), Vout: 0},
		{Txid: []byte("transaction-1-block-2"), Vout: 1},
	})
	require.NoError(t, err)
	require.Equal(t, uint(75), balance)

	balance, err = utxos.RetrieveInputsBalance([]kernel.TxInput{
		{Txid: []byte("coinbase-transaction-block-2"), Vout: 0},
		{Txid: []byte("transaction-1-block-2"), Vout: 1},
		{Txid: []byte("coinbase-transaction-block-3"), Vout: 0},
	})
	require.NoError(t, err)
	require.Equal(t, uint(125), balance)
}

func TestUTXOSet_RetrieveInputsBalanceWithInvalidInput(t *testing.T) {
	utxos := NewUTXOSet(config.NewConfig())

	utxos.OnBlockAddition(b1)
	utxos.OnBlockAddition(b2)
	utxos.OnBlockAddition(b3)

	_, err := utxos.RetrieveInputsBalance([]kernel.TxInput{
		{Txid: []byte("coinbase-transaction-block-2"), Vout: 1}, // Vout does not exist
	})
	require.Error(t, err)

	_, err = utxos.RetrieveInputsBalance([]kernel.TxInput{
		{Txid: []byte("coinbase-transaction-block-1"), Vout: 0}, // Spent output
	})
	require.Error(t, err)

	_, err = utxos.RetrieveInputsBalance([]kernel.TxInput{
		{Txid: []byte("random-id"), Vout: 0}, // Random ID
	})
	require.Error(t, err)
}

func TestUTXOSet_AddBlockWithInvalidInput(t *testing.T) {
	utxos := NewUTXOSet(config.NewConfig())

	// add block that references input not present in the UTXO set
	require.Error(t, utxos.AddBlock(b2))
}
