package blockchain

import (
	"chainnet/config"
	"chainnet/pkg/block"
	"chainnet/pkg/storage"
)

type Iterator interface {
	GetNextBlock() (*block.Block, error)
	HasNext() bool
}

type IteratorStruct struct {
	prevBlockHash []byte
	storage       storage.Storage

	cfg *config.Config
}

func NewIterator(lastBlockHash []byte, storage storage.Storage, cfg *config.Config) *IteratorStruct {
	return &IteratorStruct{
		prevBlockHash: lastBlockHash,
		storage:       storage,
		cfg:           cfg,
	}
}

func (it *IteratorStruct) GetNextBlock() (*block.Block, error) {
	block, err := it.storage.RetrieveBlockByHash(it.prevBlockHash)
	if err != nil {
		return nil, err
	}

	it.prevBlockHash = block.PrevBlockHash

	return block, err
}

func (it *IteratorStruct) HasNext() bool {
	return len(it.prevBlockHash) > 0
}
