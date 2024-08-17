package storage

import (
	"chainnet/pkg/kernel"
	"errors"
)

const (
	FirstBlockKey  = "firstblock"
	FirstHeaderKey = "firstheader"
	LastBlockKey   = "lastblock"
	LastHeaderKey  = "lastheader"
	// LastBlockHashKey is updated when persisting a new block header
	LastBlockHashKey = "lastblockhash"

	StorageObserverID = "storage-observer"
)

var ErrNotFound = errors.New("not found")

type Storage interface {
	// PersistBlock stores a new block and updates LastBlockKey
	PersistBlock(block kernel.Block) error
	// PersistHeader stores a new header and updates LastHeaderKey and LastBlockHashKey. The latter key
	// is updated in this function because as soon as the header is written the block has been commited
	// to the chain, even if the block itself has not been persisted yet. Refer to GetLastBlockHash
	// function for additional information
	PersistHeader(blockHash []byte, blockHeader kernel.BlockHeader) error
	// GetLastBlock retrieves the block information contained in LastBlockKey
	GetLastBlock() (*kernel.Block, error)
	// GetLastHeader retrieves the header of the last block. The last header represents the latest block
	// commited to the chain (the block may not be persisted yet)
	GetLastHeader() (*kernel.BlockHeader, error)
	// GetLastBlockHash retrieves the latest block added to the chain. The value retrieved in this function
	// is updated when PersistHeader is called, ensuring therefore that as soon as a header is persisted, the
	// latest block hash is persisted too (this have to do with the chain observer).
	//
	// There may be cases in which the block hash is present but the block represented has not been persisted
	// yet. This case is well known and the chain package is designed taking that fact into account.
	GetLastBlockHash() ([]byte, error)
	// GetGenesisBlock retrieves the content stored in FirstBlockKey
	GetGenesisBlock() (*kernel.Block, error)
	// GetGenesisHeader retrieves the content stored in FirstHeaderKey
	GetGenesisHeader() (*kernel.BlockHeader, error)
	// RetrieveBlockByHash retrieves the block that corresponds to the hash
	RetrieveBlockByHash(hash []byte) (*kernel.Block, error)
	// RetrieveHeaderByHash retrieves the block header that corresponds to the block hash
	RetrieveHeaderByHash(hash []byte) (*kernel.BlockHeader, error)
	// ID returns the key StorageObserverID used for running Observer code
	ID() string
	// OnBlockAddition called when a new block is added to the chain, in the case of storage must be async
	OnBlockAddition(block *kernel.Block)
	// Close finishes the connection with the DB
	Close() error
}
