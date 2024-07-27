package explorer

import (
	"chainnet/pkg/chain/iterator"
	"chainnet/pkg/kernel"
	"chainnet/pkg/storage"
	"encoding/hex"
)

type Explorer struct {
	storage storage.Storage
}

func NewExplorer(storage storage.Storage) *Explorer {
	return &Explorer{storage: storage}
}

// GetLastBlockHash returns the hash of the last block in the blockchain
func (explorer *Explorer) GetLastBlock() (*kernel.Block, error) {
	block, err := explorer.storage.GetLastBlock()
	if err != nil {
		return nil, err
	}

	// todo(): consider returning block header directly instead of the whole block
	return block, nil
}

func (explorer *Explorer) FindUnspentTransactions(pubKey string) ([]*kernel.Transaction, error) {
	return explorer.findUnspentTransactions(pubKey, iterator.NewReverseIterator(explorer.storage))
}

// findUnspentTransactions finds all unspent transaction outputs that can be unlocked with the given address. Starts
// by checking the outputs and later the inputs, this is done this way in order to follow the inverse flow
// of transactions
// todo() remove this method, we will be using findUnspentOutputs instead most likely
func (explorer *Explorer) findUnspentTransactions(pubKey string, it iterator.BlockIterator) ([]*kernel.Transaction, error) { //nolint:gocognit // ok for now
	var nextBlock *kernel.Block
	var unspentTXs []*kernel.Transaction

	spentTXOs := make(map[string][]uint)

	lastBlock, err := explorer.storage.GetLastBlock()
	if err != nil {
		return []*kernel.Transaction{}, err
	}

	// get the blockchain revIterator
	_ = it.Initialize(lastBlock.Hash)

	for it.HasNext() {
		// get the next block using the revIterator
		nextBlock, err = it.GetNextBlock()
		if err != nil {
			return []*kernel.Transaction{}, err
		}

		// skip the genesis block
		if nextBlock.IsGenesisBlock() {
			continue
		}

		// iterate through each transaction in the block
		for _, tx := range nextBlock.Transactions {
			txID := hex.EncodeToString(tx.ID)

			for outIdx, out := range tx.Vout {
				// in case is already spent, continue
				if isOutputSpent(spentTXOs, txID, uint(outIdx)) {
					continue
				}

				// check if the output can be unlocked with the given pubKey
				if out.CanBeUnlockedWith(pubKey) {
					unspentTXs = append(unspentTXs, tx)
				}
			}

			// we skip the coinbase transactions inputs
			if tx.IsCoinbase() {
				continue
			}

			// if not coinbase, iterate through inputs and save the already spent outputs
			for _, in := range tx.Vin {
				if in.CanUnlockOutputWith(pubKey) {
					inTxID := hex.EncodeToString(in.Txid)

					// mark the output as spent
					spentTXOs[inTxID] = append(spentTXOs[inTxID], in.Vout)
				}
			}
		}
	}

	// todo() may be worth to output the utxos directly instead of the whole transaction

	// return the list of unspent transactions
	return unspentTXs, nil
}

func (explorer *Explorer) FindUnspentOutputs(pubKey string) ([]kernel.UnspentOutput, error) {
	return explorer.findUnspentOutputs(pubKey, iterator.NewReverseIterator(explorer.storage))
}

// findUnspentOutputs finds all unspent outputs that can be unlocked with the given public key
func (explorer *Explorer) findUnspentOutputs(pubKey string, it iterator.BlockIterator) ([]kernel.UnspentOutput, error) { //nolint:gocognit // ok for now
	var nextBlock *kernel.Block
	unspentTXOs := []kernel.UnspentOutput{}
	spentTXOs := make(map[string][]uint)

	lastBlock, err := explorer.storage.GetLastBlock()
	if err != nil {
		return []kernel.UnspentOutput{}, err
	}

	// get the blockchain revIterator
	_ = it.Initialize(lastBlock.Hash)

	for it.HasNext() {
		// get the next block using the revIterator
		nextBlock, err = it.GetNextBlock()
		if err != nil {
			return []kernel.UnspentOutput{}, err
		}

		// skip the genesis block
		if nextBlock.IsGenesisBlock() {
			continue
		}

		// iterate through each transaction in the block
		for _, tx := range nextBlock.Transactions {
			for outIdx, out := range tx.Vout {
				// in case is already spent, continue
				if isOutputSpent(spentTXOs, string(tx.ID), uint(outIdx)) {
					continue
				}

				// check if the output can be unlocked with the given pubKey
				if out.CanBeUnlockedWith(pubKey) {
					unspentTXOs = append(unspentTXOs, kernel.UnspentOutput{
						TxID:   tx.ID,
						OutIdx: uint(outIdx),
						Output: out,
					})
				}
			}

			// we skip the coinbase transactions inputs
			if tx.IsCoinbase() {
				continue
			}

			// if not coinbase, iterate through inputs and save the already spent outputs
			for _, in := range tx.Vin {
				if in.CanUnlockOutputWith(pubKey) {
					// mark the output as spent
					spentTXOs[string(in.Txid)] = append(spentTXOs[string(in.Txid)], in.Vout)
				}
			}
		}
	}

	// return the list of unspent transactions
	return unspentTXOs, nil
}

func (explorer *Explorer) CalculateAddressBalance(pubKey string) (uint, error) {
	unspentTXs, err := explorer.FindUnspentTransactionsOutputs(pubKey)
	if err != nil {
		return 0, err
	}

	return retrieveBalanceFrom(unspentTXs), nil
}

func (explorer *Explorer) FindAmountSpendableOutputs(pubKey string, amount uint) (uint, map[string][]uint, error) {
	unspentOutputs := make(map[string][]uint)
	unspentTXs, err := explorer.FindUnspentTransactions(pubKey)
	if err != nil {
		return uint(0), unspentOutputs, err
	}

	accumulated := uint(0)

	// retrieve all unspent transactions and sum them
	for _, tx := range unspentTXs {
		txID := hex.EncodeToString(tx.ID)

		for outIdx, out := range tx.Vout {
			if out.CanBeUnlockedWith(pubKey) {
				accumulated += out.Amount
				unspentOutputs[txID] = append(unspentOutputs[txID], uint(outIdx))

				// return once we reached the required amount
				if accumulated >= amount {
					return accumulated, unspentOutputs, nil
				}
			}
		}
	}

	// there is a chance that we don't have enough amount for this pubKey
	return accumulated, unspentOutputs, nil
}

func (explorer *Explorer) FindUnspentTransactionsOutputs(pubKey string) ([]kernel.TxOutput, error) {
	unspentTransactions, err := explorer.FindUnspentTransactions(pubKey)
	if err != nil {
		return []kernel.TxOutput{}, err
	}

	return explorer.findUnspentTransactionsOutputs(pubKey, unspentTransactions)
}

func (explorer *Explorer) findUnspentTransactionsOutputs(pubKey string, unspentTransactions []*kernel.Transaction) ([]kernel.TxOutput, error) {
	var utxos []kernel.TxOutput

	for _, tx := range unspentTransactions {
		for _, out := range tx.Vout {
			if out.CanBeUnlockedWith(pubKey) {
				utxos = append(utxos, out)
			}
		}
	}

	return utxos, nil
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

// retrieveBalanceFrom calculates the total amount of unspent transactions outputs
func retrieveBalanceFrom(utxos []kernel.TxOutput) uint {
	accumulated := uint(0)

	for _, out := range utxos {
		accumulated += out.Amount
	}

	return accumulated
}
