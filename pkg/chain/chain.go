package blockchain

import (
	"chainnet/config"
	exp "chainnet/pkg/chain/explorer"
	"chainnet/pkg/consensus"
	"chainnet/pkg/kernel"
	"chainnet/pkg/script"
	"chainnet/pkg/storage"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
)

// AdjustDifficultyHeight adjusts difficulty every 2016 blocks (~2 weeks)
const AdjustDifficultyHeight = 2016

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

// AddBlock takes transactions and creates a new kernel, then it calculates the hash and nonce for the kernel
// and persists it in the storage. It also updates the last kernel hash and the chain
func (bc *Blockchain) AddBlock(transactions []*kernel.Transaction) (*kernel.Block, error) {
	var hash []byte
	var nonce uint
	var newBlock *kernel.Block

	numBlocks, err := bc.storage.NumberOfBlocks()
	if err != nil {
		return &kernel.Block{}, err
	}

	// if no blocks exist, create genesis kernel
	if numBlocks == 0 {
		newBlock = kernel.NewGenesisBlock(transactions)
	}

	// if blocks exist, create new kernel tied to the previous
	if numBlocks > 0 {
		newBlock = kernel.NewBlock(transactions, bc.lastBlockHash)
	}

	// todo() this will probably be validated in the miner/node level
	err = bc.validator.ValidateBlock(newBlock)
	if err != nil {
		return &kernel.Block{}, fmt.Errorf("error validating block %s: %w", newBlock.Hash, err)
	}

	newBlock.Target = bc.cfg.DifficultyPoW

	// calculate hash and nonce for the kernel
	// todo() not really a very good approach, clear it or split in more funcs: Mine for example
	for {
		newBlock.Timestamp = time.Now().Unix()
		// calculate until the max nonce, if does not match, try again with different timestamp
		hash, nonce, err = bc.consensus.CalculateBlockHash(newBlock)
		if err == nil {
			newBlock.SetHashAndNonce(hash, nonce)
			break
		}
	}

	// persist kernel and update information
	err = bc.storage.PersistBlock(*newBlock)
	if err != nil {
		return &kernel.Block{}, err
	}

	bc.lastBlockHash = newBlock.Hash
	bc.Chain = append(bc.Chain, string(newBlock.Hash))

	return newBlock, nil
}

func (bc *Blockchain) NewCoinbaseTransaction(to string) (*kernel.Transaction, error) {
	tx := kernel.NewCoinbaseTransaction(to)

	txHash, err := bc.consensus.CalculateTxHash(tx)
	if err != nil {
		return nil, err
	}

	tx.SetID(txHash)

	return tx, nil
}

func (bc *Blockchain) NewTransaction(from, to string, amount uint) (*kernel.Transaction, error) {
	var inputs []kernel.TxInput
	var outputs []kernel.TxOutput

	// todo() delete this aberration once explorer & NewTransaction is more developed (should be created from wallet, not here)
	explorer := exp.NewExplorer(bc.storage)

	acc, validOutputs, err := explorer.FindAmountSpendableOutputs(from, amount)
	if err != nil {
		return &kernel.Transaction{}, err
	}

	if acc < amount {
		return &kernel.Transaction{}, fmt.Errorf("not enough funds for transaction from %s, expected tx amount: %d, actual balance: %d", from, amount, acc)
	}

	// build a list of inputs
	// todo() move to another function
	var txID []byte
	for txid, outs := range validOutputs {
		txID, err = hex.DecodeString(txid)
		if err != nil {
			return &kernel.Transaction{}, err
		}

		for _, out := range outs {
			input := kernel.NewInput(txID, out, from, from)
			inputs = append(inputs, input)
		}
	}

	// build a list of outputs
	outputs = append(outputs, kernel.NewOutput(amount, script.P2PK, to))

	// add the spare change in a different transaction
	if acc > amount {
		outputs = append(outputs, kernel.NewOutput(acc-amount, script.P2PK, from))
	}

	tx := kernel.NewTransaction(inputs, outputs)

	txHash, err := bc.consensus.CalculateTxHash(tx)
	if err != nil {
		return &kernel.Transaction{}, err
	}

	tx.SetID(txHash)

	// todo() this will probably be validated in the miner/node level
	err = bc.validator.ValidateTx(tx)
	if err != nil {
		return &kernel.Transaction{}, fmt.Errorf("error validating transaction %s: %w", tx.ID, err)
	}

	return tx, nil
}

func (bc *Blockchain) MineBlock(transactions []*kernel.Transaction) *kernel.Block {
	newBlock := kernel.NewBlock(transactions, bc.lastBlockHash)

	return newBlock
}

func (bc *Blockchain) GetBlock(hash string) (*kernel.Block, error) {
	return bc.storage.RetrieveBlockByHash([]byte(hash))
}

func (bc *Blockchain) GetLastBlockHash() []byte {
	return bc.lastBlockHash
}
