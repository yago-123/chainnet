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

func TestBlockchain_findUnspentTransactions(t *testing.T) {
	// set up genesis kernel with coinbase transaction
	coinbaseTx := NewCoinbaseTransaction("pubKey-1")
	coinbaseTx.SetID([]byte("coinbase-transaction-genesis-id"))
	genesisBlock := NewGenesisBlock([]*Transaction{coinbaseTx})
	genesisBlock.SetHashAndNonce([]byte("genesis-kernel-hash"), 1)

	// set up kernel 1 with one coinbase transaction
	coinbaseTx1 := NewCoinbaseTransaction("pubKey-2")
	coinbaseTx1.SetID([]byte("coinbase-transaction-kernel-1-id"))
	block1 := NewBlock([]*Transaction{coinbaseTx1}, genesisBlock.Hash)
	block1.SetHashAndNonce([]byte("kernel-hash-1"), 1)

	// set up kernel 2 with one coinbase transaction and one regular transaction
	coinbaseTx2 := NewCoinbaseTransaction("pubKey-3")
	coinbaseTx2.SetID([]byte("coinbase-transaction-kernel-2-id"))
	regularTx := NewTransaction(
		[]TxInput{
			NewInput([]byte("coinbase-transaction-kernel-1-id"), 0, "pubKey-2", "pubKey-2"),
		},
		[]TxOutput{
			NewOutput(2, script.P2PK, "pubKey-3"),
			NewOutput(3, script.P2PK, "pubKey-4"),
			NewOutput(44, script.P2PK, "pubKey-5"),
			NewOutput(1, script.P2PK, "pubKey-2"),
		})
	regularTx.SetID([]byte("regular-transaction-kernel-2-id"))
	block2 := NewBlock([]*Transaction{coinbaseTx2, regularTx}, block1.Hash)
	block2.SetHashAndNonce([]byte("kernel-hash-2"), 1)

	// set up kernel 3 with one coinbase transaction and two regular transactions
	coinbaseTx3 := NewCoinbaseTransaction("pubKey-4")
	coinbaseTx3.SetID([]byte("coinbase-transaction-kernel-3-id"))
	regularTx2 := NewTransaction(
		[]TxInput{
			NewInput([]byte("regular-transaction-kernel-2-id"), 1, "pubKey-4", "pubKey-4"),
			NewInput([]byte("regular-transaction-kernel-2-id"), 2, "pubKey-5", "pubKey-5"),
		},
		[]TxOutput{
			NewOutput(4, script.P2PK, "pubKey-2"),
			NewOutput(2, script.P2PK, "pubKey-3"),
			NewOutput(41, script.P2PK, "pubKey-4"),
		},
	)
	regularTx2.SetID([]byte("regular-transaction-kernel-3-id"))
	regularTx3 := NewTransaction(
		[]TxInput{
			NewInput([]byte("regular-transaction-kernel-2-id"), 0, "pubKey-3", "pubKey-3"),
		},
		[]TxOutput{
			NewOutput(1, script.P2PK, "pubKey-6"),
			NewOutput(1, script.P2PK, "pubKey-3"),
		},
	)
	regularTx3.SetID([]byte("regular-transaction-2-kernel-3-id"))
	block3 := NewBlock([]*Transaction{coinbaseTx3, regularTx2, regularTx3}, block2.Hash)
	block3.SetHashAndNonce([]byte("kernel-hash-3"), 1)

	// set up kernel 4 with one coinbase transaction
	coinbaseTx4 := NewCoinbaseTransaction("pubKey-7")
	coinbaseTx4.SetID([]byte("coinbase-transaction-kernel-4-id"))
	block4 := NewBlock([]*Transaction{coinbaseTx4}, block3.Hash)
	block4.SetHashAndNonce([]byte("kernel-hash-4"), 1)

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
			Return(block4, nil)

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
			Return(block4, nil).
			Once()
		reverseIterator.
			On("GetNextBlock").
			Return(block3, nil).
			Once()
		reverseIterator.
			On("GetNextBlock").
			Return(block2, nil).
			Once()
		reverseIterator.
			On("GetNextBlock").
			Return(block1, nil).
			Once()
		reverseIterator.
			On("GetNextBlock").
			Return(genesisBlock, nil).
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
