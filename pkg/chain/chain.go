package blockchain

import (
	"chainnet/pkg/block"
	"chainnet/pkg/config"
	"chainnet/pkg/consensus"
	"chainnet/pkg/storage"
	"encoding/hex"
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

func (bc *Blockchain) AddBlock(transactions []*block.Transaction) (*block.Block, error) {
	var newBlock *block.Block

	numBlocks, err := bc.storage.NumberOfBlocks()
	if err != nil {
		return &block.Block{}, err
	}

	// if no blocks exist, create genesis block
	if numBlocks == 0 {
		newBlock = block.NewBlock(transactions, []byte{})
	}

	// if blocks exist, create new block tied to the previous
	if numBlocks > 0 {
		newBlock = block.NewBlock(transactions, bc.lastBlockHash)
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

// FindUnspentTransactions finds all unspent transaction outputs that can be unlocked with the given address. Starts
// by checking the outputs and later the inputs, this is done this way in order to follow the inverse flow
// of transactions
func (bc *Blockchain) FindUnspentTransactions(address string) []*block.Transaction {
	var unspentTXs []*block.Transaction
	spentTXOs := make(map[string][]int)

	// Get the blockchain iterator
	bciterator := bc.CreateIterator()

	for bciterator.HasNext() {
		// Get the next block using the iterator
		confirmedBlock, err := bciterator.GetNextBlock()
		if err != nil {
			// todo() handle error
		}

		// iterate through each transaction in the block
		for _, tx := range confirmedBlock.Transactions {
			txID := hex.EncodeToString(tx.ID)

			for outIdx, out := range tx.Vout {
				// in case is already spent, continue
				if isOutputSpent(spentTXOs, txID, outIdx) {
					continue
				}

				// check if the output can be unlocked with the given address
				if out.CanBeUnlockedWith(address) {
					unspentTXs = append(unspentTXs, tx)
				}
			}

			// if not coinbase, iterate through inputs and save the already spent outputs
			if !tx.IsCoinbase() {
				for _, in := range tx.Vin {
					if in.CanUnlockOutputWith(address) {
						inTxID := hex.EncodeToString(in.Txid)

						// mark the output as spent
						spentTXOs[inTxID] = append(spentTXOs[inTxID], in.Vout)
					}
				}
			}
		}
	}

	// Return the list of unspent transactions
	return unspentTXs
}

func isOutputSpent(spentTXOs map[string][]int, txID string, outIdx int) bool {
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

func (bc *Blockchain) FindUTXO(address string) []block.TxOutput {
	var UTXOs []block.TxOutput
	unspentTransactions := bc.FindUnspentTransactions(address)

	for _, tx := range unspentTransactions {
		for _, out := range tx.Vout {
			if out.CanBeUnlockedWith(address) {
				UTXOs = append(UTXOs, out)
			}
		}
	}

	return UTXOs
}

func (bc *Blockchain) GetBlock(hash string) (*block.Block, error) {
	return bc.storage.RetrieveBlockByHash([]byte(hash))
}

func (bc *Blockchain) CreateIterator() Iterator {
	return NewIterator(bc.lastBlockHash, bc.storage, bc.cfg)
}
