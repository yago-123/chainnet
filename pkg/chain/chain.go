package blockchain

import (
	"chainnet/config"
	"chainnet/pkg/chain/observer"
	"chainnet/pkg/consensus"
	"chainnet/pkg/consensus/util"
	"chainnet/pkg/crypto/hash"
	"chainnet/pkg/kernel"
	"chainnet/pkg/storage"
	"fmt"

	"github.com/sirupsen/logrus"
)

type Blockchain struct {
	lastBlockHash []byte
	lastHeight    uint
	headers       map[string]kernel.BlockHeader
	// blockTxsBloomFilter map[string]string

	storage   storage.Storage
	validator consensus.HeavyValidator
	subject   *observer.SubjectObserver

	logger *logrus.Logger
	cfg    *config.Config
}

func NewBlockchain(cfg *config.Config, storage storage.Storage, hasher hash.Hashing, validator consensus.HeavyValidator, subject *observer.SubjectObserver) (*Blockchain, error) {
	var err error
	var lastHeight uint
	var lastBlockHash []byte

	headers := make(map[string]kernel.BlockHeader)

	// retrieve the last header stored
	lastHeader, err := storage.GetLastHeader()
	if err != nil {
		return nil, fmt.Errorf("error retrieving last header: %w", err)
	}

	// if exists a last header, sync the actual status of the chain
	if !lastHeader.IsEmpty() {
		// specify the current height
		lastHeight = lastHeader.Height + 1

		// get the last block hash by hashing the latest block header
		lastBlockHash, err = util.CalculateBlockHash(lastHeader, hasher)
		if err != nil {
			return nil, fmt.Errorf("error retrieving last block hash: %w", err)
		}

		// reload the headers into memory
		headers, err = reconstructHeaders(lastBlockHash)
		if err != nil {
			return nil, fmt.Errorf("error reconstructing headers: %w", err)
		}
	}

	return &Blockchain{
		lastBlockHash: lastBlockHash,
		lastHeight:    lastHeight,
		headers:       headers,
		storage:       storage,
		validator:     validator,
		subject:       subject,
		logger:        cfg.Logger,
		cfg:           cfg,
	}, nil
}

func (bc *Blockchain) Sync() {
	// if there is latest block (node has been started before)
	// - retrieve latest block hash
	// - compare height and hash
	// - decide with the other peers next steps (download next headers or execute some conflict resolution)

	// if there is not latest block (node is new)
	// - ask for headers
	// - download & verify headers
	// - start IBD (Initial Block Download): download block from each header
	// 		- validate each block
}

func (bc *Blockchain) AddBlock(block *kernel.Block) error {
	if err := bc.validator.ValidateBlock(block); err != nil {
		return fmt.Errorf("block validation failed: %w", err)
	}

	// persist block header, once the header has been persisted the block has been commited to the chain
	if err := bc.storage.PersistHeader(block.Hash, *block.Header); err != nil {
		return fmt.Errorf("block header persistence failed: %w", err)
	}

	// STARTING FROM HERE: the code can fail without becoming an issue, the header has been already commited
	// no need to store the block itself, will be commited to storage as part of the observer call

	// update the last block and save the block header
	bc.lastBlockHash = block.Hash
	bc.headers[string(block.Hash)] = *block.Header

	// notify observers of a new block added
	bc.subject.NotifyBlockAdded(block)

	return nil
}

// GetLastBlockHash returns the latest block hash
func (bc *Blockchain) GetLastBlockHash() []byte {
	return bc.lastBlockHash
}

// GetLastHeight returns the latest block height
func (bc *Blockchain) GetLastHeight() uint {
	return bc.lastHeight
}

func reconstructHeaders(latestBlockHash []byte) (map[string]kernel.BlockHeader, error) {
	return map[string]kernel.BlockHeader{}, nil
}
