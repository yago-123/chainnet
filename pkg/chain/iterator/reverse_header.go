package iterator

import (
	"chainnet/pkg/kernel"
	"chainnet/pkg/storage"
)

// ReverseHeaderIterator 
type ReverseHeaderIterator struct {
	prevBlockHash []byte
	storage       storage.Storage
}

func NewReverseHeaderIterator(storage storage.Storage) *ReverseHeaderIterator {
	return &ReverseHeaderIterator{
		storage: storage,
	}
}

func (it *ReverseHeaderIterator) Initialize(reference []byte) error {
	it.prevBlockHash = reference
	return nil
}

func (it *ReverseHeaderIterator) GetNextBlock() (*kernel.BlockHeader, error) {
	header, err := it.storage.RetrieveHeaderByHash(it.prevBlockHash)
	if err != nil {
		return nil, err
	}

	it.prevBlockHash = header.PrevBlockHash

	return header, err
}

func (it *ReverseHeaderIterator) HasNext() bool {
	return len(it.prevBlockHash) > 0
}
