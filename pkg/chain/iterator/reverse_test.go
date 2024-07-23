package iterator //nolint:testpackage // don't create separate package for tests

import (
	"chainnet/pkg/consensus/miner"
	"chainnet/pkg/kernel"
	mockStorage "chainnet/tests/mocks/storage"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReverseIterator(t *testing.T) {
	// set up genesis kernel with coinbase transaction
	coinbaseTx := kernel.NewCoinbaseTransaction("address-1", miner.InitialCoinbaseReward, 0)
	coinbaseTx.SetID([]byte("coinbase-transaction-genesis-id"))
	blockHeader := kernel.NewBlockHeader([]byte{}, 0, []byte{}, 0, []byte{}, 0, 0)
	blockHeader.SetNonce(1)
	genesisBlock := kernel.NewGenesisBlock(blockHeader, []*kernel.Transaction{coinbaseTx}, []byte("genesis-kernel-hash"))

	// set up kernel 1 with one coinbase transaction
	coinbaseTx2 := kernel.NewCoinbaseTransaction("address-2", miner.InitialCoinbaseReward, 0)
	coinbaseTx2.SetID([]byte("coinbase-transaction-kernel-1-id"))
	blockHeader = kernel.NewBlockHeader([]byte{}, 0, []byte{}, 0, genesisBlock.Hash, 0, 0)
	blockHeader.SetNonce(1)
	block1 := kernel.NewBlock(blockHeader, []*kernel.Transaction{coinbaseTx2}, []byte("kernel-hash-1"))

	// set up kernel 2 with one coinbase transaction and one regular transaction
	coinbaseTx3 := kernel.NewCoinbaseTransaction("address-3", miner.InitialCoinbaseReward, 0)
	coinbaseTx3.SetID([]byte("coinbase-transaction-kernel-2-id"))
	regularTx := kernel.NewTransaction(
		[]kernel.TxInput{
			{
				Txid:      []byte("coinbase-transaction-kernel-1-id"),
				Vout:      0,
				ScriptSig: "ScriptSig",
			},
		},
		[]kernel.TxOutput{
			{Amount: 1, ScriptPubKey: "ScriptPubKey"},
		})
	regularTx.SetID([]byte("regular-tx-2-id"))
	blockHeader = kernel.NewBlockHeader([]byte{}, 0, []byte{}, 0, block1.Hash, 0, 0)
	blockHeader.SetNonce(1)
	block2 := kernel.NewBlock(blockHeader, []*kernel.Transaction{coinbaseTx3, regularTx}, []byte("kernel-hash-2"))

	// set up the storage mock with the corresponding returns
	storage := &mockStorage.MockStorage{}
	storage.
		On("RetrieveBlockByHash", block2.Hash).
		Return(block2, nil)
	storage.
		On("RetrieveBlockByHash", block1.Hash).
		Return(block1, nil)
	storage.
		On("RetrieveBlockByHash", genesisBlock.Hash).
		Return(genesisBlock, nil)

	// check that the iterator works as expected
	reverseIterator := NewReverseIterator(storage)

	// initialize iterator with the last kernel hash
	err := reverseIterator.Initialize(block2.Hash)
	require.NoError(t, err)

	// check if we have next element and retrieve kernel 2
	assert.True(t, reverseIterator.HasNext(), "error checking if next kernel exists")
	b, err := reverseIterator.GetNextBlock()
	require.NoError(t, err)
	assert.Equal(t, []byte("kernel-hash-2"), b.Hash, "failure retrieving kernel 2")

	// check if we have next element and retrieve kernel 1
	assert.True(t, reverseIterator.HasNext(), "error checking if next kernel exists")
	b, err = reverseIterator.GetNextBlock()
	require.NoError(t, err)
	assert.Equal(t, []byte("kernel-hash-1"), b.Hash, "failure retrieving kernel 1")

	// check if we have next element and retrieve genesis kernel
	assert.True(t, reverseIterator.HasNext(), "error checking if next kernel exists")
	b, err = reverseIterator.GetNextBlock()
	require.NoError(t, err)
	assert.Equal(t, []byte("genesis-kernel-hash"), b.Hash, "failure retrieving genesis kernel")

	// verify that we don't have more elements available
	assert.False(t, reverseIterator.HasNext(), "more elements were found when the chain must have ended")
}

func TestReverseIteratorWithOnlyGenesisBlock(t *testing.T) {
	// set up genesis kernel with coinbase transaction
	coinbaseTx := kernel.NewCoinbaseTransaction("address-1", miner.InitialCoinbaseReward, 0)
	coinbaseTx.SetID([]byte("coinbase-genesis-kernel-id"))
	blockHeader := kernel.NewBlockHeader([]byte{}, 0, []byte{}, 0, []byte{}, 0, 0)
	blockHeader.SetNonce(1)
	genesisBlock := kernel.NewGenesisBlock(blockHeader, []*kernel.Transaction{coinbaseTx}, []byte("genesis-kernel-hash"))

	// set up the storage mock with the corresponding returns
	storage := &mockStorage.MockStorage{}
	storage.
		On("RetrieveBlockByHash", genesisBlock.Hash).
		Return(genesisBlock, nil)

	// check that the iterator works as expected
	reverseIterator := NewReverseIterator(storage)

	// initialize iterator with the last kernel hash
	err := reverseIterator.Initialize(genesisBlock.Hash)
	require.NoError(t, err)

	// check if we have next element and retrieve genesis kernel
	assert.True(t, reverseIterator.HasNext(), "error checking if next kernel exists")
	b, err := reverseIterator.GetNextBlock()
	require.NoError(t, err)
	assert.Equal(t, []byte("genesis-kernel-hash"), b.Hash, "failure retrieving genesis kernel")

	// verify that we don't have more elements available
	assert.False(t, reverseIterator.HasNext(), "more elements were found when the chain must have ended")
}
