package storage

import "chainnet/pkg/kernel"

type Storage interface {
	NumberOfBlocks() (uint, error)

	PersistBlock(block kernel.Block) error
	PersistHeader(blockHash []byte, blockHeader kernel.BlockHeader) error

	GetLastBlock() (*kernel.Block, error)
	GetLastHeader() (*kernel.BlockHeader, error)

	GetGenesisBlock() (*kernel.Block, error)
	GetGenesisHeader() (*kernel.BlockHeader, error)

	RetrieveBlockByHash(hash []byte) (*kernel.Block, error)
	RetrieveHeaderByHash(hash []byte) (*kernel.BlockHeader, error)

	Close() error
}
