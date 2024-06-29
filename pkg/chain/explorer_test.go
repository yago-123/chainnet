package blockchain

import (
	"chainnet/config"
	"chainnet/pkg/chain/iterator"
	. "chainnet/pkg/kernel"
	"chainnet/pkg/script"
	mockIterator "chainnet/tests/mocks/chain/iterator"
	mockConsensus "chainnet/tests/mocks/consensus"
	mockStorage "chainnet/tests/mocks/storage"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"testing"
)

// set up genesis kernel with coinbase transaction
var GenesisBlock = &Block{
	Timestamp: 0,
	Transactions: []*Transaction{
		{
			ID: []byte("coinbase-transaction-genesis-id"),
			Vin: []TxInput{
				NewCoinbaseInput(),
			},
			Vout: []TxOutput{
				NewOutput(COINBASE_AMOUNT, script.P2PK, "pubKey-1"),
			},
		},
	},
	PrevBlockHash: []byte{},
	Nonce:         1,
	Hash:          []byte("genesis-kernel-hash"),
}

// set up kernel 1 with one coinbase transaction
var Block1 = &Block{
	Timestamp: 0,
	Transactions: []*Transaction{
		{
			ID: []byte("coinbase-transaction-kernel-1-id"),
			Vin: []TxInput{
				NewCoinbaseInput(),
			},
			Vout: []TxOutput{
				NewOutput(COINBASE_AMOUNT, script.P2PK, "pubKey-2"),
			},
		},
	},
	PrevBlockHash: GenesisBlock.Hash,
	Nonce:         1,
	Hash:          []byte("kernel-hash-1"),
}

// set up kernel 2 with one coinbase transaction and one regular transaction
var Block2 = &Block{
	Timestamp: 0,
	Transactions: []*Transaction{
		{
			ID: []byte("coinbase-transaction-kernel-2-id"),
			Vin: []TxInput{
				NewCoinbaseInput(),
			},
			Vout: []TxOutput{
				NewOutput(COINBASE_AMOUNT, script.P2PK, "pubKey-3"),
			},
		},
		{
			ID: []byte("regular-transaction-kernel-2-id"),
			Vin: []TxInput{
				NewInput([]byte("coinbase-transaction-kernel-1-id"), 0, "pubKey-2", "pubKey-2"),
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
	Hash:          []byte("kernel-hash-2"),
}

// set up kernel 3 with one coinbase transaction and two regular transactions
var Block3 = &Block{
	Timestamp: 0,
	Transactions: []*Transaction{
		{
			ID: []byte("coinbase-transaction-kernel-3-id"),
			Vin: []TxInput{
				NewCoinbaseInput(),
			},
			Vout: []TxOutput{
				NewOutput(COINBASE_AMOUNT, script.P2PK, "pubKey-4"),
			},
		},
		{
			ID: []byte("regular-transaction-kernel-3-id"),
			Vin: []TxInput{
				NewInput([]byte("regular-transaction-kernel-2-id"), 1, "pubKey-4", "pubKey-4"),
				NewInput([]byte("regular-transaction-kernel-2-id"), 2, "pubKey-5", "pubKey-5"),
			},
			Vout: []TxOutput{
				NewOutput(4, script.P2PK, "pubKey-2"),
				NewOutput(2, script.P2PK, "pubKey-3"),
				NewOutput(41, script.P2PK, "pubKey-4"),
			},
		},
		{
			ID: []byte("regular-transaction-2-kernel-3-id"),
			Vin: []TxInput{
				NewInput([]byte("regular-transaction-kernel-2-id"), 0, "pubKey-3", "pubKey-3"),
			},
			Vout: []TxOutput{
				NewOutput(1, script.P2PK, "pubKey-6"),
				NewOutput(1, script.P2PK, "pubKey-3"),
			},
		},
	},
	PrevBlockHash: Block2.Hash,
	Nonce:         1,
	Hash:          []byte("kernel-hash-3"),
}

// set up kernel 4 with one coinbase transaction
var Block4 = &Block{
	Timestamp: 0,
	Transactions: []*Transaction{
		{
			ID: []byte("coinbase-transaction-kernel-4-id"),
			Vin: []TxInput{
				NewCoinbaseInput(),
			},
			Vout: []TxOutput{
				NewOutput(COINBASE_AMOUNT, script.P2PK, "pubKey-7"),
			},
		},
	},
	PrevBlockHash: Block3.Hash,
	Nonce:         1,
	Hash:          []byte("kernel-hash-4"),
}

func TestBlockchain_findUnspentTransactions(t *testing.T) {

	storageInstance := &mockStorage.MockStorage{}
	bc := NewBlockchain(
		config.NewConfig(logrus.New(), 1, 1, ""),
		&mockConsensus.MockConsensus{},
		storageInstance,
	)
	bc.lastBlockHash = []byte("kernel-hash-4")

	restartedMockIterator := func() iterator.Iterator {
		reverseIterator := &mockIterator.MockIterator{}

		storageInstance.
			On("GetLastBlock").
			Return(Block4, nil)

		reverseIterator.
			On("Initialize", []byte("kernel-hash-4")).
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

	explorer := NewExplorer(storageInstance)

	txs, err := explorer.findUnspentTransactions("pubKey-1", restartedMockIterator())
	assert.NoError(t, err)
	assert.Equal(t, 0, len(txs))

	txs, err = explorer.findUnspentTransactions("pubKey-2", restartedMockIterator())
	assert.NoError(t, err)
	assert.Equal(t, 2, len(txs))
	assert.Equal(t, []byte("regular-transaction-kernel-3-id"), txs[0].ID)
	assert.Equal(t, []byte("regular-transaction-kernel-2-id"), txs[1].ID)

	txs, err = explorer.findUnspentTransactions("pubKey-3", restartedMockIterator())
	assert.NoError(t, err)
	assert.Equal(t, 3, len(txs))
	assert.Equal(t, []byte("regular-transaction-kernel-3-id"), txs[0].ID)
	assert.Equal(t, []byte("regular-transaction-2-kernel-3-id"), txs[1].ID)
	assert.Equal(t, []byte("coinbase-transaction-kernel-2-id"), txs[2].ID)

	txs, err = explorer.findUnspentTransactions("pubKey-4", restartedMockIterator())
	assert.NoError(t, err)
	assert.Equal(t, 2, len(txs))
	assert.Equal(t, "coinbase-transaction-kernel-3-id", string(txs[0].ID))
	assert.Equal(t, "regular-transaction-kernel-3-id", string(txs[1].ID))

	txs, err = explorer.findUnspentTransactions("pubKey-5", restartedMockIterator())
	assert.NoError(t, err)
	assert.Equal(t, 0, len(txs))

	txs, err = explorer.findUnspentTransactions("pubKey-6", restartedMockIterator())
	assert.NoError(t, err)
	assert.Equal(t, 1, len(txs))
	assert.Equal(t, []byte("regular-transaction-2-kernel-3-id"), txs[0].ID)

	txs, err = explorer.findUnspentTransactions("pubKey-7", restartedMockIterator())
	assert.NoError(t, err)
	assert.Equal(t, 1, len(txs))
	assert.Equal(t, []byte("coinbase-transaction-kernel-4-id"), txs[0].ID)

}

func TestBlockchain_FindAmountSpendableOutput(t *testing.T) {

}

func TestBlockchain_FindUTXO(t *testing.T) {

}

func TestBlockchain_isOutputSpent(t *testing.T) {
	spentTXOs := make(map[string][]uint)

	spentTXOs["tx-spent-1"] = []uint{0, 1}
	spentTXOs["tx-spent-2"] = []uint{0}
	spentTXOs["tx-spent-3"] = []uint{3}

	assert.Equal(t, true, isOutputSpent(spentTXOs, "tx-spent-1", 0))
	assert.Equal(t, true, isOutputSpent(spentTXOs, "tx-spent-1", 1))
	assert.Equal(t, true, isOutputSpent(spentTXOs, "tx-spent-2", 0))
	assert.Equal(t, true, isOutputSpent(spentTXOs, "tx-spent-3", 3))

	assert.Equal(t, false, isOutputSpent(spentTXOs, "tx-spent-1", 2))
	assert.Equal(t, false, isOutputSpent(spentTXOs, "tx-spent-2", 1))
	assert.Equal(t, false, isOutputSpent(spentTXOs, "tx-spent-3", 0))
}

func TestBlockchain_retrieveBalanceFrom(t *testing.T) {
	utxos := []TxOutput{
		NewOutput(1, script.P2PK, "random-1"),
		NewOutput(2, script.P2PK, "random-2"),
		NewOutput(100, script.P2PK, "random-3"),
	}

	assert.Equal(t, uint(103), retrieveBalanceFrom(utxos))
}
