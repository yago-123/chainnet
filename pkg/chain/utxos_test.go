package blockchain //nolint:testpackage // don't create separate package for tests

import (
	"chainnet/pkg/kernel"
	"chainnet/pkg/script"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

var b1 = &kernel.Block{
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

var b2 = &kernel.Block{
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

var b3 = &kernel.Block{
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
	utxos := NewUTXOSet()

	utxos.AddBlock(b1)
	utxos.AddBlock(b2)
	utxos.AddBlock(b3)

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

func TestUTXOSet_AddBlockWithInvalidInput(t *testing.T) {
	utxos := NewUTXOSet()

	// add block that references input not in the UTXO set
	require.Error(t, utxos.AddBlock(b2))

}
