package blockchain

import (
	"chainnet/block"
	"chainnet/config"
	"chainnet/consensus"
	"chainnet/storage"
	"github.com/sirupsen/logrus"
)

type Blockchain struct {
	Chain  []string
	Blocks map[string]*block.Block

	consensus   consensus.Consensus
	persistence storage.Storage

	logger *logrus.Logger
	cfg    *config.Config
}

func NewBlockchain(cfg *config.Config, consensus consensus.Consensus, persistence storage.Storage) *Blockchain {
	bc := &Blockchain{
		Chain:       []string{},
		Blocks:      make(map[string]*block.Block),
		consensus:   consensus,
		persistence: persistence,
		logger:      cfg.Logger,
		cfg:         cfg,
	}

	numBlocks, err := persistence.NumberOfBlocks()
	if err != nil {
		logrus.Panicf("Error creating genesis block: %s", err)
	}

	if numBlocks == 0 {
		bc.NewGenesisBlock()
	}
	return bc
}

func (bc *Blockchain) NewGenesisBlock() *block.Block {
	newBlock := block.NewBlock("Genesis block", []byte{})

	hash, nonce := bc.consensus.Calculate(newBlock)
	newBlock.SetHashAndNonce(hash, nonce)

	bc.Blocks[string(newBlock.Hash)] = newBlock
	bc.Chain = append(bc.Chain, string(newBlock.Hash))

	return newBlock
}

func (bc *Blockchain) AddBlock(data string) *block.Block {
	prevBlock := bc.Blocks[bc.Chain[len(bc.Chain)-1]]
	newBlock := block.NewBlock(data, prevBlock.Hash)

	hash, nonce := bc.consensus.Calculate(newBlock)
	newBlock.SetHashAndNonce(hash, nonce)

	bc.Blocks[string(newBlock.Hash)] = newBlock
	bc.Chain = append(bc.Chain, string(newBlock.Hash))

	return newBlock
}
