package explorer //nolint:testpackage // don't create separate package for tests

import (
	"chainnet/pkg/chain/iterator"
	. "chainnet/pkg/kernel"
	"chainnet/pkg/script"
	mockIterator "chainnet/tests/mocks/chain/iterator"
	mockStorage "chainnet/tests/mocks/storage"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// set up genesis block with coinbase transaction
var GenesisBlock = &Block{ //nolint:gochecknoglobals // data that is used across all test funcs
	Timestamp: 0,
	Transactions: []*Transaction{
		{
			ID: []byte("coinbase-transaction-genesis-id"),
			Vin: []TxInput{
				NewCoinbaseInput(),
			},
			Vout: []TxOutput{
				NewOutput(CoinbaseReward, script.P2PK, "pubKey-1"),
			},
		},
	},
	PrevBlockHash: []byte{},
	Nonce:         1,
	Hash:          []byte("genesis-block-hash"),
}

// set up block 1 with one coinbase transaction
var Block1 = &Block{ //nolint:gochecknoglobals // data that is used across all test funcs
	Timestamp: 0,
	Transactions: []*Transaction{
		{
			ID: []byte("coinbase-transaction-block-1-id"),
			Vin: []TxInput{
				NewCoinbaseInput(),
			},
			Vout: []TxOutput{
				NewOutput(CoinbaseReward, script.P2PK, "pubKey-2"),
			},
		},
	},
	PrevBlockHash: GenesisBlock.Hash,
	Nonce:         1,
	Hash:          []byte("block-hash-1"),
}

// set up block 2 with one coinbase transaction and one regular transaction
var Block2 = &Block{ //nolint:gochecknoglobals // data that is used across all test funcs
	Timestamp: 0,
	Transactions: []*Transaction{
		{
			ID: []byte("coinbase-transaction-block-2-id"),
			Vin: []TxInput{
				NewCoinbaseInput(),
			},
			Vout: []TxOutput{
				NewOutput(CoinbaseReward, script.P2PK, "pubKey-3"),
			},
		},
		{
			ID: []byte("regular-transaction-block-2-id"),
			Vin: []TxInput{
				NewInput([]byte("coinbase-transaction-block-1-id"), 0, "pubKey-2", "pubKey-2"),
			},
			Vout: []TxOutput{
				NewOutput(2, script.P2PK, "pubKey-3"),
				NewOutput(3, script.P2PK, "pubKey-4"),
				NewOutput(44, script.P2PK, "pubKey-5"),
				NewOutput(1, script.P2PK, "pubKey-2"),
			},
		},
	},
	PrevBlockHash: Block1.Hash,
	Nonce:         1,
	Hash:          []byte("block-hash-2"),
}

// set up block 3 with one coinbase transaction and two regular transactions
var Block3 = &Block{ //nolint:gochecknoglobals // data that is used across all test funcs
	Timestamp: 0,
	Transactions: []*Transaction{
		{
			ID: []byte("coinbase-transaction-block-3-id"),
			Vin: []TxInput{
				NewCoinbaseInput(),
			},
			Vout: []TxOutput{
				NewOutput(CoinbaseReward, script.P2PK, "pubKey-4"),
			},
		},
		{
			ID: []byte("regular-transaction-block-3-id"),
			Vin: []TxInput{
				NewInput([]byte("regular-transaction-block-2-id"), 1, "pubKey-4", "pubKey-4"),
				NewInput([]byte("regular-transaction-block-2-id"), 2, "pubKey-5", "pubKey-5"),
			},
			Vout: []TxOutput{
				NewOutput(4, script.P2PK, "pubKey-2"),
				NewOutput(2, script.P2PK, "pubKey-3"),
				NewOutput(41, script.P2PK, "pubKey-4"),
			},
		},
		{
			ID: []byte("regular-transaction-2-block-3-id"),
			Vin: []TxInput{
				NewInput([]byte("regular-transaction-block-2-id"), 0, "pubKey-3", "pubKey-3"),
			},
			Vout: []TxOutput{
				NewOutput(1, script.P2PK, "pubKey-6"),
				NewOutput(1, script.P2PK, "pubKey-3"),
			},
		},
	},
	PrevBlockHash: Block2.Hash,
	Nonce:         1,
	Hash:          []byte("block-hash-3"),
}

// set up block 4 with one coinbase transaction
var Block4 = &Block{ //nolint:gochecknoglobals // data that is used across all test funcs
	Timestamp: 0,
	Transactions: []*Transaction{
		{
			ID: []byte("coinbase-transaction-block-4-id"),
			Vin: []TxInput{
				NewCoinbaseInput(),
			},
			Vout: []TxOutput{
				NewOutput(CoinbaseReward, script.P2PK, "pubKey-7"),
			},
		},
	},
	PrevBlockHash: Block3.Hash,
	Nonce:         1,
	Hash:          []byte("block-hash-4"),
}

func TestBlockchain_findUnspentTransactions(t *testing.T) {
	storageInstance := &mockStorage.MockStorage{}
	storageInstance.
		On("GetLastBlock").
		Return(Block4, nil)

	explorer := NewExplorer(storageInstance)

	// todo(): split each pubKey check into a separate test so is more descriptive
	txs, err := explorer.findUnspentTransactions("pubKey-1", initializeMockedChain())
	require.NoError(t, err)
	assert.Empty(t, txs)

	txs, err = explorer.findUnspentTransactions("pubKey-2", initializeMockedChain())
	require.NoError(t, err)
	assert.Len(t, txs, 2)
	assert.Equal(t, []byte("regular-transaction-block-3-id"), txs[0].ID)
	assert.Equal(t, []byte("regular-transaction-block-2-id"), txs[1].ID)

	txs, err = explorer.findUnspentTransactions("pubKey-3", initializeMockedChain())
	require.NoError(t, err)
	assert.Len(t, txs, 3)
	assert.Equal(t, []byte("regular-transaction-block-3-id"), txs[0].ID)
	assert.Equal(t, []byte("regular-transaction-2-block-3-id"), txs[1].ID)
	assert.Equal(t, []byte("coinbase-transaction-block-2-id"), txs[2].ID)

	txs, err = explorer.findUnspentTransactions("pubKey-4", initializeMockedChain())
	require.NoError(t, err)
	assert.Len(t, txs, 2)
	assert.Equal(t, "coinbase-transaction-block-3-id", string(txs[0].ID))
	assert.Equal(t, "regular-transaction-block-3-id", string(txs[1].ID))

	txs, err = explorer.findUnspentTransactions("pubKey-5", initializeMockedChain())
	require.NoError(t, err)
	assert.Empty(t, txs)

	txs, err = explorer.findUnspentTransactions("pubKey-6", initializeMockedChain())
	require.NoError(t, err)
	assert.Len(t, txs, 1)
	assert.Equal(t, []byte("regular-transaction-2-block-3-id"), txs[0].ID)

	txs, err = explorer.findUnspentTransactions("pubKey-7", initializeMockedChain())
	require.NoError(t, err)
	assert.Len(t, txs, 1)
	assert.Equal(t, []byte("coinbase-transaction-block-4-id"), txs[0].ID)
}

func TestBlockchain_findUnspentOutputs(t *testing.T) {
	storageInstance := &mockStorage.MockStorage{}
	storageInstance.
		On("GetLastBlock").
		Return(Block4, nil)

	explorer := NewExplorer(storageInstance)

	// todo(): split each pubKey check into a separate test so is more descriptive
	// todo(): add additional checks for the other fields in the TxOutput struct
	utxo, err := explorer.findUnspentOutputs("pubKey-1", initializeMockedChain())
	require.NoError(t, err)
	assert.Empty(t, utxo)

	utxo, err = explorer.findUnspentOutputs("pubKey-2", initializeMockedChain())
	require.NoError(t, err)
	assert.Len(t, utxo, 2)
	assert.Equal(t, []byte("regular-transaction-block-3-id"), utxo[0].TxID)
	assert.Equal(t, []byte("regular-transaction-block-2-id"), utxo[1].TxID)

	utxo, err = explorer.findUnspentOutputs("pubKey-3", initializeMockedChain())
	require.NoError(t, err)
	assert.Len(t, utxo, 3)
	assert.Equal(t, []byte("regular-transaction-block-3-id"), utxo[0].TxID)
	assert.Equal(t, []byte("regular-transaction-2-block-3-id"), utxo[1].TxID)
	assert.Equal(t, []byte("coinbase-transaction-block-2-id"), utxo[2].TxID)

	utxo, err = explorer.findUnspentOutputs("pubKey-4", initializeMockedChain())
	require.NoError(t, err)
	assert.Len(t, utxo, 2)
	assert.Equal(t, "coinbase-transaction-block-3-id", string(utxo[0].TxID))
	assert.Equal(t, "regular-transaction-block-3-id", string(utxo[1].TxID))

	utxo, err = explorer.findUnspentOutputs("pubKey-5", initializeMockedChain())
	require.NoError(t, err)
	assert.Empty(t, utxo)

	utxo, err = explorer.findUnspentOutputs("pubKey-6", initializeMockedChain())
	require.NoError(t, err)
	assert.Len(t, utxo, 1, utxo)
	assert.Equal(t, []byte("regular-transaction-2-block-3-id"), utxo[0].TxID)

	utxo, err = explorer.findUnspentOutputs("pubKey-7", initializeMockedChain())
	require.NoError(t, err)
	assert.Len(t, utxo, 1)
	assert.Equal(t, []byte("coinbase-transaction-block-4-id"), utxo[0].TxID)
}

func TestBlockchain_FindAmountSpendableOutput(_ *testing.T) {

}

func TestBlockchain_FindUTXO(_ *testing.T) {

}

func TestBlockchain_isOutputSpent(t *testing.T) {
	spentTXOs := make(map[string][]uint)

	spentTXOs["tx-spent-1"] = []uint{0, 1}
	spentTXOs["tx-spent-2"] = []uint{0}
	spentTXOs["tx-spent-3"] = []uint{3}

	assert.True(t, isOutputSpent(spentTXOs, "tx-spent-1", 0))
	assert.True(t, isOutputSpent(spentTXOs, "tx-spent-1", 1))
	assert.True(t, isOutputSpent(spentTXOs, "tx-spent-2", 0))
	assert.True(t, isOutputSpent(spentTXOs, "tx-spent-3", 3))

	assert.False(t, isOutputSpent(spentTXOs, "tx-spent-1", 2))
	assert.False(t, isOutputSpent(spentTXOs, "tx-spent-2", 1))
	assert.False(t, isOutputSpent(spentTXOs, "tx-spent-3", 0))
}

func TestBlockchain_retrieveBalanceFrom(t *testing.T) {
	utxos := []TxOutput{
		NewOutput(1, script.P2PK, "random-1"),
		NewOutput(2, script.P2PK, "random-2"),
		NewOutput(100, script.P2PK, "random-3"),
	}

	assert.Equal(t, uint(103), retrieveBalanceFrom(utxos))
}

func initializeMockedChain() iterator.Iterator {
	reverseIterator := &mockIterator.MockIterator{}

	reverseIterator.
		On("Initialize", []byte("block-hash-4")).
		Return(nil)

	reverseIterator.
		On("HasNext").
		Return(true).
		Times(5)
	reverseIterator.
		On("HasNext").
		Return(false).
		Once()

	reverseIterator.
		On("GetNextBlock").
		Return(Block4, nil).
		Once()
	reverseIterator.
		On("GetNextBlock").
		Return(Block3, nil).
		Once()
	reverseIterator.
		On("GetNextBlock").
		Return(Block2, nil).
		Once()
	reverseIterator.
		On("GetNextBlock").
		Return(Block1, nil).
		Once()
	reverseIterator.
		On("GetNextBlock").
		Return(GenesisBlock, nil).
		Once()

	return reverseIterator
}
