package blockchain //nolint:testpackage // don't create separate package for tests

import (
	"chainnet/config"
	. "chainnet/pkg/kernel" //nolint:revive // it's fine to use dot imports in tests
	"chainnet/pkg/script"

	"github.com/stretchr/testify/require"

	mockConsensus "chainnet/tests/mocks/consensus"
	mockStorage "chainnet/tests/mocks/storage"
	mockUtil "chainnet/tests/mocks/util"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
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
	require.NoError(t, err, "errors while adding genesis blockAdded")
	assert.Empty(t, blockAdded.PrevBlockHash, "genesis blockAdded contains previous blockAdded hash when it shouldn't")
	assert.Equal(t, []byte("genesis-kernel-hash"), blockAdded.Hash, "blockAdded hash incorrect")
	assert.Equal(t, uint(1), blockAdded.Nonce, "blockAdded nonce incorrect")
	assert.Equal(t, []byte("genesis-kernel-hash"), bc.lastBlockHash, "last blockAdded hash in blockchain not updated")
	assert.Len(t, bc.Chain, 1, "blockchain chain length not updated")
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
	require.NoError(t, err, "errors while adding genesis blockAdded")
	assert.Equal(t, []byte("genesis-kernel-hash"), blockAdded.PrevBlockHash, "blockAdded contains previous blockAdded hash when it shouldn't")
	assert.Equal(t, []byte("second-kernel-hash"), blockAdded.Hash, "blockAdded hash incorrect")
	assert.Equal(t, uint(1), blockAdded.Nonce, "blockAdded nonce incorrect")
	assert.Equal(t, []byte("second-kernel-hash"), bc.lastBlockHash, "last blockAdded hash in blockchain not updated")
	assert.Len(t, bc.Chain, 2, "blockchain chain length not updated")
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
	require.NoError(t, err, "errors while adding genesis blockAdded")
	assert.Equal(t, []byte("second-kernel-hash"), blockAdded.PrevBlockHash, "blockAdded contains previous blockAdded hash when it shouldn't")
	assert.Equal(t, []byte("third-kernel-hash"), blockAdded.Hash, "blockAdded hash incorrect")
	assert.Equal(t, uint(1), blockAdded.Nonce, "blockAdded nonce incorrect")
	assert.Equal(t, []byte("third-kernel-hash"), bc.lastBlockHash, "last blockAdded hash in blockchain not updated")
	assert.Len(t, bc.Chain, 3, "blockchain chain length not updated")
	assert.Equal(t, "third-kernel-hash", bc.Chain[2], "blockchain chain not updated with new blockAdded hash")
}

func TestBlockchain_AddBlockWithErrors(_ *testing.T) {

}

func TestBlockchain_AddBlockWithInvalidTransaction(_ *testing.T) {

}
