package blockchain

import (
	"chainnet/config"
	"chainnet/pkg/chain/iterator"
	. "chainnet/pkg/kernel"
	"chainnet/pkg/script"

	mockIterator "chainnet/tests/mocks/chain/iterator"
	mockConsensus "chainnet/tests/mocks/consensus"
	mockStorage "chainnet/tests/mocks/storage"
	mockUtil "chainnet/tests/mocks/util"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestBlockchain_AddBlockWithoutErrors(t *testing.T) {
	bc := NewBlockchain(
		config.NewConfig(logrus.New(), 1, 1, ""),
		&mockConsensus.MockConsensus{},
		&mockStorage.MockStorage{},
	)

	coinbaseTx := []*Transaction{
		{
			ID: nil,
			Vin: []TxInput{
				NewCoinbaseInput(),
			},
			Vout: []TxOutput{
				NewCoinbaseOutput(script.P2PK, "pubKey"),
			},
		},
	}

	secondTx := []*Transaction{
		{
			ID: []byte("second-tx-id"),
			Vin: []TxInput{
				NewInput([]byte("random"), 0, "random-script-sig", "random-script-sig"),
			},
			Vout: []TxOutput{
				NewOutput(100, script.P2PK, "random-pub-key"),
				NewOutput(200, script.P2PK, "random-script-sig"),
			},
		},
	}

	thirdTx := []*Transaction{
		{
			ID: []byte("third-tx-id"),
			Vin: []TxInput{
				NewInput([]byte("random"), 0, "random-script-sig", "random-script-sig"),
			},
			Vout: []TxOutput{
				NewOutput(101, script.P2PK, "random-pub-key-3"),
				NewOutput(201, script.P2PK, "random-pub-key-4"),
			},
		},
	}

	// setup the return values for the internal AddBlock calls
	bc.storage.(*mockStorage.MockStorage).
		On("NumberOfBlocks").
		Return(uint(0), nil).Once()
	bc.storage.(*mockStorage.MockStorage).
		On("PersistBlock", mockUtil.MatchByPreviousBlock([]byte{})).
		Return(nil)
	bc.consensus.(*mockConsensus.MockConsensus).
		On("CalculateBlockHash", mockUtil.MatchByPreviousBlockPointer([]byte{})).
		Return([]byte("genesis-kernel-hash"), uint(1), nil)

	// add genesis kernel
	blockAdded, err := bc.AddBlock(coinbaseTx)

	// check that the blockAdded has been added correctly
	assert.Equal(t, nil, err, "errors while adding genesis blockAdded")
	assert.Equal(t, 0, len(blockAdded.PrevBlockHash), "genesis blockAdded contains previous blockAdded hash when it shouldn't")
	assert.Equal(t, []byte("genesis-kernel-hash"), blockAdded.Hash, "blockAdded hash incorrect")
	assert.Equal(t, uint(1), blockAdded.Nonce, "blockAdded nonce incorrect")
	assert.Equal(t, []byte("genesis-kernel-hash"), bc.lastBlockHash, "last blockAdded hash in blockchain not updated")
	assert.Equal(t, 1, len(bc.Chain), "blockchain chain length not updated")
	assert.Equal(t, "genesis-kernel-hash", bc.Chain[0], "blockchain chain not updated with new blockAdded hash")

	// setup the return values for the internal AddBlock calls
	bc.storage.(*mockStorage.MockStorage).
		On("NumberOfBlocks").
		Return(uint(1), nil).Once()
	bc.storage.(*mockStorage.MockStorage).
		On("PersistBlock", mockUtil.MatchByPreviousBlock([]byte("genesis-kernel-hash"))).
		Return(nil)
	bc.consensus.(*mockConsensus.MockConsensus).
		On("CalculateBlockHash", mockUtil.MatchByPreviousBlockPointer([]byte("genesis-kernel-hash"))).
		Return([]byte("second-kernel-hash"), uint(1), nil)

	// add another kernel
	blockAdded, err = bc.AddBlock(secondTx)

	// check that the blockAdded has been added correctly
	assert.Equal(t, nil, err, "errors while adding genesis blockAdded")
	assert.Equal(t, []byte("genesis-kernel-hash"), blockAdded.PrevBlockHash, "blockAdded contains previous blockAdded hash when it shouldn't")
	assert.Equal(t, []byte("second-kernel-hash"), blockAdded.Hash, "blockAdded hash incorrect")
	assert.Equal(t, uint(1), blockAdded.Nonce, "blockAdded nonce incorrect")
	assert.Equal(t, []byte("second-kernel-hash"), bc.lastBlockHash, "last blockAdded hash in blockchain not updated")
	assert.Equal(t, 2, len(bc.Chain), "blockchain chain length not updated")
	assert.Equal(t, "second-kernel-hash", bc.Chain[1], "blockchain chain not updated with new blockAdded hash")

	// setup the return values for the internal AddBlock calls
	bc.storage.(*mockStorage.MockStorage).
		On("NumberOfBlocks").
		Return(uint(2), nil).Once()
	bc.storage.(*mockStorage.MockStorage).
		On("PersistBlock", mockUtil.MatchByPreviousBlock([]byte("second-kernel-hash"))).
		Return(nil)
	bc.consensus.(*mockConsensus.MockConsensus).
		On("CalculateBlockHash", mockUtil.MatchByPreviousBlockPointer([]byte("second-kernel-hash"))).
		Return([]byte("third-kernel-hash"), uint(1), nil)

	// add another kernel
	blockAdded, err = bc.AddBlock(thirdTx)

	// check that the blockAdded has been added correctly
	assert.Equal(t, nil, err, "errors while adding genesis blockAdded")
	assert.Equal(t, []byte("second-kernel-hash"), blockAdded.PrevBlockHash, "blockAdded contains previous blockAdded hash when it shouldn't")
	assert.Equal(t, []byte("third-kernel-hash"), blockAdded.Hash, "blockAdded hash incorrect")
	assert.Equal(t, uint(1), blockAdded.Nonce, "blockAdded nonce incorrect")
	assert.Equal(t, []byte("third-kernel-hash"), bc.lastBlockHash, "last blockAdded hash in blockchain not updated")
	assert.Equal(t, 3, len(bc.Chain), "blockchain chain length not updated")
	assert.Equal(t, "third-kernel-hash", bc.Chain[2], "blockchain chain not updated with new blockAdded hash")
}

func TestBlockchain_AddBlockWithErrors(t *testing.T) {

}

func TestBlockchain_AddBlockWithInvalidTransaction(t *testing.T) {

}

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

/////

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
