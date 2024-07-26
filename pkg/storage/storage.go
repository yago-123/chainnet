package storage

import "chainnet/pkg/kernel"

type Storage interface {
	NumberOfBlocks() (uint, error)

	PersistBlock(block kernel.Block) error
	PersistHeader(blockHash []byte, blockHeader kernel.BlockHeader) error

	GetLastBlock() (*kernel.Block, error)
	GetGenesisBlock() (*kernel.Block, error)
	RetrieveBlockByHash(hash []byte) (*kernel.Block, error)
	RetrieveHeaderByHash(hash []byte) (*kernel.BlockHeader, error)
	Close() error
}
