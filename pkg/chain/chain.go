package blockchain

import (
	"chainnet/config"
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

	logger *logrus.Logger
	cfg    *config.Config
}

func NewBlockchain(cfg *config.Config, consensus consensus.Consensus, storage storage.Storage, validator consensus.HeavyValidator) *Blockchain {
	bc := &Blockchain{
		Chain:     []string{},
		consensus: consensus,
		storage:   storage,
		validator: validator,
		logger:    cfg.Logger,
		cfg:       cfg,
	}

	return bc
}

func (bc *Blockchain) AddBlock(block *kernel.Block) error {
	if err := bc.validator.ValidateBlock(block); err != nil {
		return fmt.Errorf("block validation failed: %w", err)
	}

	// propagate to utxo?

	// propagate to mempool?

	// propagate to miner?

	// propagate to network?

	// notify peers about new block?

	return nil
}

func (bc *Blockchain) GetBlock(hash string) (*kernel.Block, error) {
	return bc.storage.RetrieveBlockByHash([]byte(hash))
}

func (bc *Blockchain) GetLastBlockHash() []byte {
	return bc.lastBlockHash
}
