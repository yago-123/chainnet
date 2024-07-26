package blockchain

import (
	"chainnet/config"
	"chainnet/pkg/chain/observer"
	"chainnet/pkg/consensus"
	"chainnet/pkg/kernel"
	"chainnet/pkg/storage"
	"fmt"

	"github.com/sirupsen/logrus"
)

type Blockchain struct {
	lastBlockHash       []byte
	headers             map[string]kernel.BlockHeader
	blockTxsBloomFilter map[string]string

	consensus consensus.Consensus
	storage   storage.Storage
	validator consensus.HeavyValidator
	subject   *observer.SubjectObserver

	logger *logrus.Logger
	cfg    *config.Config
}

func NewBlockchain(cfg *config.Config, consensus consensus.Consensus, storage storage.Storage, validator consensus.HeavyValidator, subject *observer.SubjectObserver) (*Blockchain, error) {
	lastBlock, err := storage.GetLastBlock()
	if err != nil {
		return nil, err
	}

	return &Blockchain{
		lastBlockHash: lastBlock.Hash,
		headers:       map[string]kernel.BlockHeader{},
		consensus:     consensus,
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
	return bc.headers[string(bc.lastBlockHash)].Height
}
