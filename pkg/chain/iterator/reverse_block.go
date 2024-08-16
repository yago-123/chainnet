package iterator

import (
	"chainnet/pkg/kernel"
	"chainnet/pkg/storage"
)

// ReverseBlockIterator
type ReverseBlockIterator struct {
	prevBlockHash []byte
	store         storage.Storage
}

func NewReverseBlockIterator(store storage.Storage) *ReverseBlockIterator {
	return &ReverseBlockIterator{
		store: store,
	}
}

func (it *ReverseBlockIterator) Initialize(reference []byte) error {
	it.prevBlockHash = reference
	return nil
}

func (it *ReverseBlockIterator) GetNextBlock() (*kernel.Block, error) {
	block, err := it.store.RetrieveBlockByHash(it.prevBlockHash)
	if err != nil {
		return nil, err
	}

	it.prevBlockHash = block.Header.PrevBlockHash

	return block, err
}

func (it *ReverseBlockIterator) HasNext() bool {
	return len(it.prevBlockHash) > 0
}
