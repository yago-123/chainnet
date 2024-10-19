package iterator

import (
	"github.com/yago-123/chainnet/pkg/kernel"
	"github.com/yago-123/chainnet/pkg/storage"
)

// ReverseHeaderIterator
type ReverseHeaderIterator struct {
	prevHeaderHash []byte
	store          storage.Storage
}

func NewReverseHeaderIterator(store storage.Storage) *ReverseHeaderIterator {
	return &ReverseHeaderIterator{
		store: store,
	}
}

func (it *ReverseHeaderIterator) Initialize(reference []byte) error {
	it.prevHeaderHash = reference
	return nil
}

func (it *ReverseHeaderIterator) GetNextHeader() (*kernel.BlockHeader, error) {
	header, err := it.store.RetrieveHeaderByHash(it.prevHeaderHash)
	if err != nil {
		return nil, err
	}

	it.prevHeaderHash = header.PrevBlockHash

	return header, err
}

func (it *ReverseHeaderIterator) HasNext() bool {
	return len(it.prevHeaderHash) > 0
}
