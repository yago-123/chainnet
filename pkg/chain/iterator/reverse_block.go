package iterator

import (
	"chainnet/pkg/kernel"
	"chainnet/pkg/storage"
)

// ReverseBlockIterator
type ReverseBlockIterator struct {
	prevBlockHash []byte
	storage       storage.Storage
}

func NewReverseBlockIterator(storage storage.Storage) *ReverseBlockIterator {
	return &ReverseBlockIterator{
		storage: storage,
	}
}

func (it *ReverseBlockIterator) Initialize(reference []byte) error {
	it.prevBlockHash = reference
	return nil
}

func (it *ReverseBlockIterator) GetNextBlock() (*kernel.Block, error) {
	block, err := it.storage.RetrieveBlockByHash(it.prevBlockHash)
	if err != nil {
		return nil, err
	}

	it.prevBlockHash = block.Header.PrevBlockHash

	return block, err
}

func (it *ReverseBlockIterator) HasNext() bool {
	return len(it.prevBlockHash) > 0
}
