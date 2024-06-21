package iterator

import (
	"chainnet/pkg/block"
	"chainnet/pkg/storage"
)

// ReverseIterator
type ReverseIterator struct {
	prevBlockHash []byte
	storage       storage.Storage
}

func NewReverseIterator(storage storage.Storage) *ReverseIterator {
	return &ReverseIterator{
		storage: storage,
	}
}

func (it *ReverseIterator) Initialize(reference []byte) error {
	it.prevBlockHash = reference
	return nil
}

func (it *ReverseIterator) GetNextBlock() (*block.Block, error) {
	block, err := it.storage.RetrieveBlockByHash(it.prevBlockHash)
	if err != nil {
		return nil, err
	}

	it.prevBlockHash = block.PrevBlockHash

	return block, err
}

func (it *ReverseIterator) HasNext() bool {
	return len(it.prevBlockHash) > 0
}
