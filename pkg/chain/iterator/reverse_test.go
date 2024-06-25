package iterator

import (
	"chainnet/pkg/kernel"
	mockStorage "chainnet/tests/mocks/storage"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestReverseIterator(t *testing.T) {
	// set up genesis kernel with coinbase transaction
	coinbaseTx := kernel.NewCoinbaseTransaction("address-1")
	coinbaseTx.SetID([]byte("coinbase-transaction-genesis-id"))
	genesisBlock := kernel.NewGenesisBlock([]*kernel.Transaction{coinbaseTx})
	genesisBlock.SetHashAndNonce([]byte("genesis-kernel-hash"), 1)

	// set up kernel 1 with one coinbase transaction
	coinbaseTx2 := kernel.NewCoinbaseTransaction("address-2")
	coinbaseTx2.SetID([]byte("coinbase-transaction-kernel-1-id"))
	block1 := kernel.NewBlock([]*kernel.Transaction{coinbaseTx2}, genesisBlock.Hash)
	block1.SetHashAndNonce([]byte("kernel-hash-1"), 1)

	// set up kernel 2 with one coinbase transaction and one regular transaction
	coinbaseTx3 := kernel.NewCoinbaseTransaction("address-3")
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
	block2 := kernel.NewBlock([]*kernel.Transaction{coinbaseTx3, regularTx}, block1.Hash)
	block2.SetHashAndNonce([]byte("kernel-hash-2"), 1)

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
	assert.NoError(t, err)

	// check if we have next element and retrieve kernel 2
	assert.Equal(t, true, reverseIterator.HasNext(), "error checking if next kernel exists")
	b, err := reverseIterator.GetNextBlock()
	assert.NoError(t, err)
	assert.Equal(t, []byte("kernel-hash-2"), b.Hash, "failure retrieving kernel 2")

	// check if we have next element and retrieve kernel 1
	assert.Equal(t, true, reverseIterator.HasNext(), "error checking if next kernel exists")
	b, err = reverseIterator.GetNextBlock()
	assert.NoError(t, err)
	assert.Equal(t, []byte("kernel-hash-1"), b.Hash, "failure retrieving kernel 1")

	// check if we have next element and retrieve genesis kernel
	assert.Equal(t, true, reverseIterator.HasNext(), "error checking if next kernel exists")
	b, err = reverseIterator.GetNextBlock()
	assert.NoError(t, err)
	assert.Equal(t, []byte("genesis-kernel-hash"), b.Hash, "failure retrieving genesis kernel")

	// verify that we don't have more elements available
	assert.Equal(t, false, reverseIterator.HasNext(), "more elements were found when the chain must have ended")
}

func TestReverseIteratorWithOnlyGenesisBlock(t *testing.T) {
	// set up genesis kernel with coinbase transaction
	coinbaseTx := kernel.NewCoinbaseTransaction("address-1")
	coinbaseTx.SetID([]byte("coinbase-genesis-kernel-id"))
	genesisBlock := kernel.NewGenesisBlock([]*kernel.Transaction{coinbaseTx})
	genesisBlock.SetHashAndNonce([]byte("genesis-kernel-hash"), 1)

	// set up the storage mock with the corresponding returns
	storage := &mockStorage.MockStorage{}
	storage.
		On("RetrieveBlockByHash", genesisBlock.Hash).
		Return(genesisBlock, nil)

	// check that the iterator works as expected
	reverseIterator := NewReverseIterator(storage)

	// initialize iterator with the last kernel hash
	err := reverseIterator.Initialize(genesisBlock.Hash)
	assert.NoError(t, err)

	// check if we have next element and retrieve genesis kernel
	assert.Equal(t, true, reverseIterator.HasNext(), "error checking if next kernel exists")
	b, err := reverseIterator.GetNextBlock()
	assert.NoError(t, err)
	assert.Equal(t, []byte("genesis-kernel-hash"), b.Hash, "failure retrieving genesis kernel")

	// verify that we don't have more elements available
	assert.Equal(t, false, reverseIterator.HasNext(), "more elements were found when the chain must have ended")
}
