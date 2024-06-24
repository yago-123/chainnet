package blockchain

import (
	"chainnet/config"
	"chainnet/pkg/block"
	"chainnet/pkg/chain/iterator"
	"chainnet/pkg/consensus"
	"chainnet/pkg/script"
	"chainnet/pkg/storage"
	"encoding/hex"
	"fmt"
	"github.com/sirupsen/logrus"
	"time"
)

// ADJUST_DIFFICULTY_HEIGHT adjusts difficulty every 2016 blocks (~2 weeks)
const ADJUST_DIFFICULTY_HEIGHT = 2016

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

// AddBlock takes transactions and creates a new block, then it calculates the hash and nonce for the block
// and persists it in the storage. It also updates the last block hash and the chain
func (bc *Blockchain) AddBlock(transactions []*block.Transaction) (*block.Block, error) {
	var newBlock *block.Block

	numBlocks, err := bc.storage.NumberOfBlocks()
	if err != nil {
		return &block.Block{}, err
	}

	// if no blocks exist, create genesis block
	if numBlocks == 0 {
		newBlock = block.NewGenesisBlock(transactions)
	}

	// if blocks exist, create new block tied to the previous
	if numBlocks > 0 {
		newBlock = block.NewBlock(transactions, bc.lastBlockHash)
	}

	newBlock.Target = bc.cfg.DifficultyPoW

	// calculate hash and nonce for the block
	// todo() not really a very good approach, clear it or split in more funcs: Mine for example
	for {
		newBlock.Timestamp = time.Now().Unix()
		// calculate until the max nonce, if does not match, try again with different timestamp
		hash, nonce, err := bc.consensus.CalculateBlockHash(newBlock)
		if err == nil {
			newBlock.SetHashAndNonce(hash[:], nonce)
			break
		}
	}

	// persist block and update information
	err = bc.storage.PersistBlock(*newBlock)
	if err != nil {
		return &block.Block{}, err
	}

	bc.lastBlockHash = newBlock.Hash
	bc.Chain = append(bc.Chain, string(newBlock.Hash))

	return newBlock, nil
}

func (bc *Blockchain) FindUnspentTransactions(address string) ([]*block.Transaction, error) {
	return bc.findUnspentTransactions(address, iterator.NewReverseIterator(bc.storage))
}

// findUnspentTransactions finds all unspent transaction outputs that can be unlocked with the given address. Starts
// by checking the outputs and later the inputs, this is done this way in order to follow the inverse flow
// of transactions
func (bc *Blockchain) findUnspentTransactions(address string, it iterator.Iterator) ([]*block.Transaction, error) {
	var unspentTXs []*block.Transaction
	spentTXOs := make(map[string][]uint)

	// get the blockchain revIterator
	_ = it.Initialize(bc.lastBlockHash)

	for it.HasNext() {
		// get the next block using the revIterator
		confirmedBlock, err := it.GetNextBlock()
		if err != nil {
			return []*block.Transaction{}, err
		}

		// skip the genesis block
		if confirmedBlock.IsGenesisBlock() {
			continue
		}

		// iterate through each transaction in the block
		for _, tx := range confirmedBlock.Transactions {
			txID := hex.EncodeToString(tx.ID)

			for outIdx, out := range tx.Vout {
				// in case is already spent, continue
				if isOutputSpent(spentTXOs, txID, uint(outIdx)) {
					continue
				}

				// check if the output can be unlocked with the given address
				if out.CanBeUnlockedWith(address) {
					unspentTXs = append(unspentTXs, tx)
				}
			}

			// we skip the coinbase transactions inputs
			if tx.IsCoinbase() {
				continue
			}

			// if not coinbase, iterate through inputs and save the already spent outputs
			for _, in := range tx.Vin {
				if in.CanUnlockOutputWith(address) {
					inTxID := hex.EncodeToString(in.Txid)

					// mark the output as spent
					spentTXOs[inTxID] = append(spentTXOs[inTxID], uint(in.Vout))
				}
			}
		}
	}

	// todo() may be worth to output the utxos directly instead of the whole transaction

	// return the list of unspent transactions
	return unspentTXs, nil
}

func (bc *Blockchain) CalculateAddressBalance(address string) (uint, error) {
	unspentTXs, err := bc.FindUnspentTransactionsOutputs(address)
	if err != nil {
		return 0, err
	}

	return retrieveBalanceFrom(unspentTXs), nil
}

func (bc *Blockchain) FindAmountSpendableOutputs(address string, amount uint) (uint, map[string][]uint, error) {
	unspentOutputs := make(map[string][]uint)
	unspentTXs, err := bc.FindUnspentTransactions(address)
	if err != nil {
		return uint(0), unspentOutputs, err
	}

	accumulated := uint(0)

	// retrieve all unspent transactions and sum them
	for _, tx := range unspentTXs {
		txID := hex.EncodeToString(tx.ID)

		for outIdx, out := range tx.Vout {
			if out.CanBeUnlockedWith(address) {
				accumulated += out.Amount
				unspentOutputs[txID] = append(unspentOutputs[txID], uint(outIdx))

				// return once we reached the required amount
				if accumulated >= amount {
					return accumulated, unspentOutputs, nil
				}
			}
		}
	}

	// there is a chance that we don't have enough amount for this address
	return accumulated, unspentOutputs, nil
}

func (bc *Blockchain) NewCoinbaseTransaction(to string) (*block.Transaction, error) {
	tx := block.NewCoinbaseTransaction(to)

	txHash, err := bc.consensus.CalculateTxHash(tx)
	if err != nil {
		return nil, err
	}

	tx.SetID(txHash[:])

	return tx, nil
}

func (bc *Blockchain) NewTransaction(from, to string, amount uint) (*block.Transaction, error) {
	var inputs []block.TxInput
	var outputs []block.TxOutput

	acc, validOutputs, err := bc.FindAmountSpendableOutputs(from, amount)
	if err != nil {
		return &block.Transaction{}, err
	}

	if acc < amount {
		return &block.Transaction{}, fmt.Errorf("not enough funds for transaction from %s, expected tx amount: %d, actual balance: %d", from, amount, acc)
	}

	// build a list of inputs
	// todo() move to another function
	for txid, outs := range validOutputs {
		txID, err := hex.DecodeString(txid)
		if err != nil {
			return &block.Transaction{}, err
		}

		for _, out := range outs {
			input := block.NewInput(txID, out, from, from)
			inputs = append(inputs, input)
		}
	}

	// build a list of outputs
	outputs = append(outputs, block.NewOutput(amount, script.P2PK, to))

	// add the spare change in a different transaction
	if acc > amount {
		outputs = append(outputs, block.NewOutput(acc-amount, script.P2PK, from))
	}

	tx := block.NewTransaction(inputs, outputs)

	txHash, err := bc.consensus.CalculateTxHash(tx)
	if err != nil {
		return &block.Transaction{}, err
	}

	tx.SetID(txHash[:])

	return tx, nil
}

func (bc *Blockchain) FindUnspentTransactionsOutputs(address string) ([]block.TxOutput, error) {
	unspentTransactions, err := bc.FindUnspentTransactions(address)
	if err != nil {
		return []block.TxOutput{}, err
	}

	return bc.findUnspentTransactionsOutputs(address, unspentTransactions)
}

func (bc *Blockchain) findUnspentTransactionsOutputs(address string, unspentTransactions []*block.Transaction) ([]block.TxOutput, error) {
	var utxos []block.TxOutput

	for _, tx := range unspentTransactions {
		for _, out := range tx.Vout {
			if out.CanBeUnlockedWith(address) {
				utxos = append(utxos, out)
			}
		}
	}

	return utxos, nil
}

func (bc *Blockchain) MineBlock(transactions []*block.Transaction) *block.Block {
	newBlock := block.NewBlock(transactions, bc.lastBlockHash)

	return newBlock
}

func (bc *Blockchain) GetBlock(hash string) (*block.Block, error) {
	return bc.storage.RetrieveBlockByHash([]byte(hash))
}

func (bc *Blockchain) GetLastBlockHash() []byte {
	return bc.lastBlockHash
}

// isOutputSpent checks if the output has been already spent by another input
func isOutputSpent(spentTXOs map[string][]uint, txID string, outIdx uint) bool {
	// check if the outputs have been already spent by an input before
	if spentOuts, spent := spentTXOs[txID]; spent {
		for _, spentOut := range spentOuts {
			// check if the output index matches
			if spentOut == outIdx {
				return true
			}
		}
	}

	return false
}

// retrieveBalanceFrom calculates the total amount of unspent transactions
func retrieveBalanceFrom(UTXOs []block.TxOutput) uint {
	accumulated := uint(0)

	for _, out := range UTXOs {
		accumulated += out.Amount
	}

	return accumulated
}
