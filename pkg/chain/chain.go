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
	Chain         []string
	lastBlockHash []byte
	headers       map[string]kernel.BlockHeader

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
		Chain:         []string{},
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

	// ... perform more checks ...

	// store block header in storage
	bc.headers[string(block.Hash)] = *block.Header

	// set last block
	bc.lastBlockHash = block.Hash

	// notify observers of new block
	bc.subject.NotifyBlockAdded(block)

	return nil
}

func (bc *Blockchain) GetBlock(hash string) (*kernel.Block, error) {
	return bc.storage.RetrieveBlockByHash([]byte(hash))
}

func (bc *Blockchain) GetLastBlockHash() []byte {
	return bc.lastBlockHash
}
