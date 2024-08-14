package iterator

import (
	"chainnet/pkg/kernel"
	"chainnet/pkg/storage"
)

// ReverseHeaderIterator
type ReverseHeaderIterator struct {
	prevHeaderHash []byte
	storage        storage.Storage
}

func NewReverseHeaderIterator(storage storage.Storage) *ReverseHeaderIterator {
	return &ReverseHeaderIterator{
		storage: storage,
	}
}

func (it *ReverseHeaderIterator) Initialize(reference []byte) error {
	it.prevHeaderHash = reference
	return nil
}

func (it *ReverseHeaderIterator) GetNextHeader() (*kernel.BlockHeader, error) {
	header, err := it.storage.RetrieveHeaderByHash(it.prevHeaderHash)
	if err != nil {
		return nil, err
	}

	it.prevHeaderHash = header.PrevBlockHash

	return header, err
}

func (it *ReverseHeaderIterator) HasNext() bool {
	return len(it.prevHeaderHash) > 0
}
