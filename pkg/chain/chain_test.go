package blockchain

import (
	"chainnet/config"
	"chainnet/pkg/block"
	"chainnet/pkg/chain/iterator"

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

	coinbaseTx := []*block.Transaction{
		{
			ID: nil,
			Vin: []block.TxInput{
				{
					Txid:      []byte{},
					Vout:      -1,
					ScriptSig: "randomSig",
				},
			},
			Vout: []block.TxOutput{
				{
					Amount:       block.COINBASE_AMOUNT,
					ScriptPubKey: "pubKey",
				},
			},
		},
	}

	secondTx := []*block.Transaction{
		{
			ID: []byte("second-tx-id"),
			Vin: []block.TxInput{
				{
					Txid:      []byte("random"),
					Vout:      0,
					ScriptSig: "random-script-sig",
				},
			},
			Vout: []block.TxOutput{
				{
					Amount:       100,
					ScriptPubKey: "random-pub-key",
				},
				{
					Amount:       200,
					ScriptPubKey: "random-pub-key-2",
				},
			},
		},
	}

	thirdTx := []*block.Transaction{
		{
			ID: []byte("third-tx-id"),
			Vin: []block.TxInput{
				{
					Txid:      []byte("random"),
					Vout:      0,
					ScriptSig: "random-script-sig",
				},
			},
			Vout: []block.TxOutput{
				{
					Amount:       101,
					ScriptPubKey: "random-pub-key-3",
				},
				{
					Amount:       201,
					ScriptPubKey: "random-pub-key-4",
				},
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
		Return([]byte("genesis-block-hash"), uint(1), nil)

	// add genesis block
	blockAdded, err := bc.AddBlock(coinbaseTx)

	// check that the blockAdded has been added correctly
	assert.Equal(t, nil, err, "errors while adding genesis blockAdded")
	assert.Equal(t, 0, len(blockAdded.PrevBlockHash), "genesis blockAdded contains previous blockAdded hash when it shouldn't")
	assert.Equal(t, []byte("genesis-block-hash"), blockAdded.Hash, "blockAdded hash incorrect")
	assert.Equal(t, uint(1), blockAdded.Nonce, "blockAdded nonce incorrect")
	assert.Equal(t, []byte("genesis-block-hash"), bc.lastBlockHash, "last blockAdded hash in blockchain not updated")
	assert.Equal(t, 1, len(bc.Chain), "blockchain chain length not updated")
	assert.Equal(t, "genesis-block-hash", bc.Chain[0], "blockchain chain not updated with new blockAdded hash")

	// setup the return values for the internal AddBlock calls
	bc.storage.(*mockStorage.MockStorage).
		On("NumberOfBlocks").
		Return(uint(1), nil).Once()
	bc.storage.(*mockStorage.MockStorage).
		On("PersistBlock", mockUtil.MatchByPreviousBlock([]byte("genesis-block-hash"))).
		Return(nil)
	bc.consensus.(*mockConsensus.MockConsensus).
		On("CalculateBlockHash", mockUtil.MatchByPreviousBlockPointer([]byte("genesis-block-hash"))).
		Return([]byte("second-block-hash"), uint(1), nil)

	// add another block
	blockAdded, err = bc.AddBlock(secondTx)

	// check that the blockAdded has been added correctly
	assert.Equal(t, nil, err, "errors while adding genesis blockAdded")
	assert.Equal(t, []byte("genesis-block-hash"), blockAdded.PrevBlockHash, "blockAdded contains previous blockAdded hash when it shouldn't")
	assert.Equal(t, []byte("second-block-hash"), blockAdded.Hash, "blockAdded hash incorrect")
	assert.Equal(t, uint(1), blockAdded.Nonce, "blockAdded nonce incorrect")
	assert.Equal(t, []byte("second-block-hash"), bc.lastBlockHash, "last blockAdded hash in blockchain not updated")
	assert.Equal(t, 2, len(bc.Chain), "blockchain chain length not updated")
	assert.Equal(t, "second-block-hash", bc.Chain[1], "blockchain chain not updated with new blockAdded hash")

	// setup the return values for the internal AddBlock calls
	bc.storage.(*mockStorage.MockStorage).
		On("NumberOfBlocks").
		Return(uint(2), nil).Once()
	bc.storage.(*mockStorage.MockStorage).
		On("PersistBlock", mockUtil.MatchByPreviousBlock([]byte("second-block-hash"))).
		Return(nil)
	bc.consensus.(*mockConsensus.MockConsensus).
		On("CalculateBlockHash", mockUtil.MatchByPreviousBlockPointer([]byte("second-block-hash"))).
		Return([]byte("third-block-hash"), uint(1), nil)

	// add another block
	blockAdded, err = bc.AddBlock(thirdTx)

	// check that the blockAdded has been added correctly
	assert.Equal(t, nil, err, "errors while adding genesis blockAdded")
	assert.Equal(t, []byte("second-block-hash"), blockAdded.PrevBlockHash, "blockAdded contains previous blockAdded hash when it shouldn't")
	assert.Equal(t, []byte("third-block-hash"), blockAdded.Hash, "blockAdded hash incorrect")
	assert.Equal(t, uint(1), blockAdded.Nonce, "blockAdded nonce incorrect")
	assert.Equal(t, []byte("third-block-hash"), bc.lastBlockHash, "last blockAdded hash in blockchain not updated")
	assert.Equal(t, 3, len(bc.Chain), "blockchain chain length not updated")
	assert.Equal(t, "third-block-hash", bc.Chain[2], "blockchain chain not updated with new blockAdded hash")
}

func TestBlockchain_AddBlockWithErrors(t *testing.T) {

}

func TestBlockchain_AddBlockWithInvalidTransaction(t *testing.T) {

}

func TestBlockchain_findUnspentTransactions(t *testing.T) {
	// set up genesis block with coinbase transaction
	coinbaseTx := block.NewCoinbaseTransaction("address-1", "data")
	coinbaseTx.SetID([]byte("coinbase-transaction-genesis-id"))
	genesisBlock := block.NewGenesisBlock([]*block.Transaction{coinbaseTx})
	genesisBlock.SetHashAndNonce([]byte("genesis-block-hash"), 1)

	// set up block 1 with one coinbase transaction
	coinbaseTx1 := block.NewCoinbaseTransaction("address-2", "data")
	coinbaseTx1.SetID([]byte("coinbase-transaction-block-1-id"))
	block1 := block.NewBlock([]*block.Transaction{coinbaseTx1}, genesisBlock.Hash)
	block1.SetHashAndNonce([]byte("block-hash-1"), 1)

	// set up block 2 with one coinbase transaction and one regular transaction
	coinbaseTx2 := block.NewCoinbaseTransaction("address-3", "data")
	coinbaseTx2.SetID([]byte("coinbase-transaction-block-2-id"))
	regularTx := block.NewTransaction(
		[]block.TxInput{
			{
				Txid:      []byte("coinbase-transaction-block-1-id"),
				Vout:      0,
				ScriptSig: "address-2",
			},
		},
		[]block.TxOutput{
			{Amount: 2, ScriptPubKey: "address3"},
			{Amount: 3, ScriptPubKey: "address4"},
			{Amount: 44, ScriptPubKey: "address5"},
			{Amount: 1, ScriptPubKey: "address2"},
		})
	regularTx.SetID([]byte("regular-transaction-block-2-id"))
	block2 := block.NewBlock([]*block.Transaction{coinbaseTx2, regularTx}, block1.Hash)
	block2.SetHashAndNonce([]byte("block-hash-2"), 1)

	// set up block 3 with one coinbase transaction and two regular transactions
	coinbaseTx3 := block.NewCoinbaseTransaction("address-4", "data")
	coinbaseTx3.SetID([]byte("coinbase-transaction-block-3-id"))
	regularTx2 := block.NewTransaction(
		[]block.TxInput{
			{
				Txid:      []byte("regular-transaction-block-2-id"),
				Vout:      1,
				ScriptSig: "address4",
			},
			{
				Txid:      []byte("regular-transaction-block-2-id"),
				Vout:      2,
				ScriptSig: "address5",
			},
		},
		[]block.TxOutput{
			{Amount: 4, ScriptPubKey: "address2"},
			{Amount: 2, ScriptPubKey: "address3"},
			{Amount: 41, ScriptPubKey: "address4"},
		},
	)
	regularTx2.SetID([]byte("regular-transaction-block-3-id"))
	regularTx3 := block.NewTransaction(
		[]block.TxInput{
			{
				Txid:      []byte("regular-transaction-block-2-id"),
				Vout:      0,
				ScriptSig: "address3",
			},
		},
		[]block.TxOutput{
			{Amount: 1, ScriptPubKey: "address6"},
			{Amount: 1, ScriptPubKey: "address3"},
		},
	)
	regularTx3.SetID([]byte("regular-transaction-2-block-3-id"))
	block3 := block.NewBlock([]*block.Transaction{coinbaseTx3, regularTx2, regularTx3}, block2.Hash)
	block3.SetHashAndNonce([]byte("block-hash-3"), 1)

	// set up block 4 with one coinbase transaction
	coinbaseTx4 := block.NewCoinbaseTransaction("address7", "data")
	coinbaseTx4.SetID([]byte("coinbase-transaction-block-4-id"))
	block4 := block.NewBlock([]*block.Transaction{coinbaseTx4}, block3.Hash)
	block4.SetHashAndNonce([]byte("block-hash-4"), 1)

	bc := NewBlockchain(
		config.NewConfig(logrus.New(), 1, 1, ""),
		&mockConsensus.MockConsensus{},
		&mockStorage.MockStorage{},
	)
	bc.lastBlockHash = []byte("block-hash-4")

	restartedMockIterator := func() iterator.Iterator {
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

	txs, err := bc.findUnspentTransactions("address1", restartedMockIterator())
	assert.NoError(t, err)
	assert.Equal(t, 0, len(txs))

	txs, err = bc.findUnspentTransactions("address2", restartedMockIterator())
	assert.NoError(t, err)
	assert.Equal(t, 2, len(txs))
	assert.Equal(t, []byte("regular-transaction-block-3-id"), txs[0].ID)
	assert.Equal(t, []byte("regular-transaction-block-2-id"), txs[1].ID)

	txs, err = bc.findUnspentTransactions("address3", restartedMockIterator())
	assert.NoError(t, err)
	assert.Equal(t, 2, len(txs))
	assert.Equal(t, []byte("regular-transaction-block-3-id"), txs[0].ID)
	assert.Equal(t, []byte("regular-transaction-2-block-3-id"), txs[1].ID)

	txs, err = bc.findUnspentTransactions("address4", restartedMockIterator())
	assert.NoError(t, err)
	assert.Equal(t, 1, len(txs))
	assert.Equal(t, []byte("regular-transaction-block-3-id"), txs[0].ID)

	txs, err = bc.findUnspentTransactions("address5", restartedMockIterator())
	assert.NoError(t, err)
	assert.Equal(t, 0, len(txs))

	txs, err = bc.findUnspentTransactions("address6", restartedMockIterator())
	assert.NoError(t, err)
	assert.Equal(t, 1, len(txs))
	assert.Equal(t, []byte("regular-transaction-2-block-3-id"), txs[0].ID)

	txs, err = bc.findUnspentTransactions("address7", restartedMockIterator())
	assert.NoError(t, err)
	assert.Equal(t, 1, len(txs))
	assert.Equal(t, []byte("coinbase-transaction-block-4-id"), txs[0].ID)

}

func TestBlockchain_FindAmountSpendableOutput(t *testing.T) {

}

func TestBlockchain_FindUTXO(t *testing.T) {

}

/////

func TestBlockchain_isOutputSpent(t *testing.T) {
	spentTXOs := make(map[string][]int)

	spentTXOs["tx-spent-1"] = []int{0, 1}
	spentTXOs["tx-spent-2"] = []int{0}
	spentTXOs["tx-spent-3"] = []int{3}

	assert.Equal(t, true, isOutputSpent(spentTXOs, "tx-spent-1", 0))
	assert.Equal(t, true, isOutputSpent(spentTXOs, "tx-spent-1", 1))
	assert.Equal(t, true, isOutputSpent(spentTXOs, "tx-spent-2", 0))
	assert.Equal(t, true, isOutputSpent(spentTXOs, "tx-spent-3", 3))

	assert.Equal(t, false, isOutputSpent(spentTXOs, "tx-spent-1", 2))
	assert.Equal(t, false, isOutputSpent(spentTXOs, "tx-spent-2", 1))
	assert.Equal(t, false, isOutputSpent(spentTXOs, "tx-spent-3", 0))
}

func TestBlockchain_retrieveBalanceFrom(t *testing.T) {
	utxos := []block.TxOutput{
		{Amount: 1, ScriptPubKey: "random-1"},
		{Amount: 2, ScriptPubKey: "random-2"},
		{Amount: 100, ScriptPubKey: "random-3"},
	}

	assert.Equal(t, 103, retrieveBalanceFrom(utxos))
}
