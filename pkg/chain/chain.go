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

// AdjustDifficultyHeight adjusts difficulty every 2016 blocks (~2 weeks)
const AdjustDifficultyHeight = 2016

// headers,

type Blockchain struct {
	Chain         []string
	lastBlockHash []byte

	consensus consensus.Consensus
	storage   storage.Storage
	validator consensus.HeavyValidator
	subject   observer.SubjectObserver

	logger *logrus.Logger
	cfg    *config.Config
}

func NewBlockchain(cfg *config.Config, consensus consensus.Consensus, storage storage.Storage, validator consensus.HeavyValidator, subject observer.SubjectObserver) *Blockchain {
	bc := &Blockchain{
		Chain:     []string{},
		consensus: consensus,
		storage:   storage,
		validator: validator,
		subject:   subject,
		logger:    cfg.Logger,
		cfg:       cfg,
	}

	return bc
}

func (bc *Blockchain) AddBlock(block *kernel.Block) error {
	if err := bc.validator.ValidateBlock(block); err != nil {
		return fmt.Errorf("block validation failed: %w", err)
	}

	// check if previous block hash is current latest block

	// check height of block

	// ... perform more checks ...

	// store block header in storage

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
