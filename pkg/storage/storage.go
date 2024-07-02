package storage

import "chainnet/pkg/kernel"

type Storage interface {
	NumberOfBlocks() (uint, error)
	PersistBlock(block kernel.Block) error
	GetLastBlock() (*kernel.Block, error)
	GetGenesisBlock() (*kernel.Block, error)
	RetrieveBlockByHash(hash []byte) (*kernel.Block, error)
	Close() error
}
