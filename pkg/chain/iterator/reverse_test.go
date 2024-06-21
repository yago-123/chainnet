package iterator

import (
	"chainnet/pkg/block"
	mockStorage "chainnet/tests/mocks/storage"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestReverseIterator(t *testing.T) {
	// set up genesis block with coinbase transaction
	coinbaseTx := block.NewCoinbaseTransaction("address-1", "data")
	coinbaseTx.SetID([]byte("coinbase-genesis-block-id"))
	genesisBlock := block.NewGenesisBlock([]*block.Transaction{coinbaseTx})
	genesisBlock.SetHashAndNonce([]byte("genesis-block-hash"), 1)

	// set up block 1 with one coinbase transaction
	coinbaseTx2 := block.NewCoinbaseTransaction("address-2", "data")
	coinbaseTx2.SetID([]byte("coinbase-block-1-id"))
	block1 := block.NewBlock([]*block.Transaction{coinbaseTx2}, genesisBlock.Hash)
	block1.SetHashAndNonce([]byte("block-hash-1"), 1)

	// set up block 2 with one coinbase transaction and one regular transaction
	coinbaseTx3 := block.NewCoinbaseTransaction("address-3", "data")
	coinbaseTx3.SetID([]byte("coinbase-block-2-id"))
	regularTx := block.NewTransaction(
		[]block.TxInput{
			{
				Txid:      []byte("coinbase-block-1-id"),
				Vout:      0,
				ScriptSig: "ScriptSig",
			},
		},
		[]block.TxOutput{
			{Amount: 1, ScriptPubKey: "ScriptPubKey"},
		})
	regularTx.SetID([]byte("regular-tx-block-2-id"))
	block2 := block.NewBlock([]*block.Transaction{coinbaseTx3, regularTx}, block1.Hash)
	block2.SetHashAndNonce([]byte("block-hash-2"), 1)

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

	// initialize iterator with the last block hash
	err := reverseIterator.Initialize(block2.Hash)
	assert.NoError(t, err)

	// check if we have next element and retrieve block 2
	assert.Equal(t, true, reverseIterator.HasNext(), "error checking if next block exists")
	b, err := reverseIterator.GetNextBlock()
	assert.NoError(t, err)
	assert.Equal(t, []byte("block-hash-2"), b.Hash, "failure retrieving block 2")

	// check if we have next element and retrieve block 1
	assert.Equal(t, true, reverseIterator.HasNext(), "error checking if next block exists")
	b, err = reverseIterator.GetNextBlock()
	assert.NoError(t, err)
	assert.Equal(t, []byte("block-hash-1"), b.Hash, "failure retrieving block 1")

	// check if we have next element and retrieve genesis block
	assert.Equal(t, true, reverseIterator.HasNext(), "error checking if next block exists")
	b, err = reverseIterator.GetNextBlock()
	assert.NoError(t, err)
	assert.Equal(t, []byte("genesis-block-hash"), b.Hash, "failure retrieving genesis block")

	// verify that we don't have more elements available
	assert.Equal(t, false, reverseIterator.HasNext(), "more elements were found when the chain must have ended")
}

func TestReverseIteratorWithOnlyGenesisBlock(t *testing.T) {
	// set up genesis block with coinbase transaction
	coinbaseTx := block.NewCoinbaseTransaction("address-1", "data")
	coinbaseTx.SetID([]byte("coinbase-genesis-block-id"))
	genesisBlock := block.NewGenesisBlock([]*block.Transaction{coinbaseTx})
	genesisBlock.SetHashAndNonce([]byte("genesis-block-hash"), 1)

	// set up the storage mock with the corresponding returns
	storage := &mockStorage.MockStorage{}
	storage.
		On("RetrieveBlockByHash", genesisBlock.Hash).
		Return(genesisBlock, nil)

	// check that the iterator works as expected
	reverseIterator := NewReverseIterator(storage)

	// initialize iterator with the last block hash
	err := reverseIterator.Initialize(genesisBlock.Hash)
	assert.NoError(t, err)

	// check if we have next element and retrieve genesis block
	assert.Equal(t, true, reverseIterator.HasNext(), "error checking if next block exists")
	b, err := reverseIterator.GetNextBlock()
	assert.NoError(t, err)
	assert.Equal(t, []byte("genesis-block-hash"), b.Hash, "failure retrieving genesis block")

	// verify that we don't have more elements available
	assert.Equal(t, false, reverseIterator.HasNext(), "more elements were found when the chain must have ended")
}
