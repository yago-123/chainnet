package blockchain

import (
	"chainnet/config"
	"chainnet/pkg/block"
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

func TestBlockchain_FindUnspentTransactions(t *testing.T) {
	_ = NewBlockchain(
		config.NewConfig(logrus.New(), 1, 1, ""),
		&mockConsensus.MockConsensus{},
		&mockStorage.MockStorage{},
	)
}

func TestBlockchain_FindAmountSpendableOutput(t *testing.T) {

}

func TestBlockchain_FindUTXOd(t *testing.T) {

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
