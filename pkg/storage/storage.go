package storage

import "chainnet/pkg/block"

type Storage interface {
	NumberOfBlocks() (uint, error)
	PersistBlock(block block.Block) error
	GetLastBlock() (*block.Block, error)
	RetrieveBlockByHash(hash []byte) (*block.Block, error)
}
