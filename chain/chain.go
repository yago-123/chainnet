package blockchain

import (
	"chainnet/block"
	"chainnet/config"
	"chainnet/consensus"
	"chainnet/storage"
	"github.com/sirupsen/logrus"
)

type Blockchain struct {
	Chain         []string
	lastBlockHash []byte

	consensus consensus.Consensus
	storage   storage.Storage

	logger *logrus.Logger
	cfg    *config.Config
}

func NewBlockchain(cfg *config.Config, consensus consensus.Consensus, persistence storage.Storage) *Blockchain {
	bc := &Blockchain{
		Chain:     []string{},
		consensus: consensus,
		storage:   persistence,
		logger:    cfg.Logger,
		cfg:       cfg,
	}

	return bc
}

func (bc *Blockchain) AddBlock(data string) (*block.Block, error) {
	var newBlock *block.Block

	numBlocks, err := bc.storage.NumberOfBlocks()
	if err != nil {
		return &block.Block{}, err
	}

	// if no blocks exist, create genesis block
	if numBlocks == 0 {
		newBlock = block.NewBlock(data, []byte{})
	}

	// if blocks exist, create new block tied to the previous
	if numBlocks > 0 {
		newBlock = block.NewBlock(data, bc.lastBlockHash)
	}

	hash, nonce := bc.consensus.Calculate(newBlock)
	newBlock.SetHashAndNonce(hash, nonce)

	// persist block and update information
	err = bc.storage.PersistBlock(*newBlock)
	if err != nil {
		return &block.Block{}, err
	}

	bc.lastBlockHash = newBlock.Hash
	bc.Chain = append(bc.Chain, string(newBlock.Hash))

	return newBlock, nil
}

func (bc *Blockchain) GetBlock(hash string) (*block.Block, error) {
	return bc.storage.RetrieveBlockByHash([]byte(hash))
}

func (bc *Blockchain) CreateIterator() Iterator {
	return NewIterator(bc.lastBlockHash, bc.storage, bc.cfg)
}
