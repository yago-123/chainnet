package blockchain

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

func (explorer *Explorer) FindUnspentTransactions(address string) ([]*kernel.Transaction, error) {
	return explorer.findUnspentTransactions(address, iterator.NewReverseIterator(explorer.storage))
}

// findUnspentTransactions finds all unspent transaction outputs that can be unlocked with the given address. Starts
// by checking the outputs and later the inputs, this is done this way in order to follow the inverse flow
// of transactions
func (explorer *Explorer) findUnspentTransactions(address string, it iterator.Iterator) ([]*kernel.Transaction, error) {
	var unspentTXs []*kernel.Transaction
	spentTXOs := make(map[string][]uint)

	lastBlock, err := explorer.storage.GetLastBlock()
	if err != nil {
		return []*kernel.Transaction{}, err
	}

	// get the blockchain revIterator
	_ = it.Initialize(lastBlock.Hash)

	for it.HasNext() {
		// get the next kernel using the revIterator
		confirmedBlock, err := it.GetNextBlock()
		if err != nil {
			return []*kernel.Transaction{}, err
		}

		// skip the genesis kernel
		if confirmedBlock.IsGenesisBlock() {
			continue
		}

		// iterate through each transaction in the kernel
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

func (explorer *Explorer) CalculateAddressBalance(address string) (uint, error) {
	unspentTXs, err := explorer.FindUnspentTransactionsOutputs(address)
	if err != nil {
		return 0, err
	}

	return retrieveBalanceFrom(unspentTXs), nil
}

func (explorer *Explorer) FindAmountSpendableOutputs(address string, amount uint) (uint, map[string][]uint, error) {
	unspentOutputs := make(map[string][]uint)
	unspentTXs, err := explorer.FindUnspentTransactions(address)
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

func (explorer *Explorer) FindUnspentTransactionsOutputs(address string) ([]kernel.TxOutput, error) {
	unspentTransactions, err := explorer.FindUnspentTransactions(address)
	if err != nil {
		return []kernel.TxOutput{}, err
	}

	return explorer.findUnspentTransactionsOutputs(address, unspentTransactions)
}

func (explorer *Explorer) findUnspentTransactionsOutputs(address string, unspentTransactions []*kernel.Transaction) ([]kernel.TxOutput, error) {
	var utxos []kernel.TxOutput

	for _, tx := range unspentTransactions {
		for _, out := range tx.Vout {
			if out.CanBeUnlockedWith(address) {
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

// retrieveBalanceFrom calculates the total amount of unspent transactions
func retrieveBalanceFrom(UTXOs []kernel.TxOutput) uint {
	accumulated := uint(0)

	for _, out := range UTXOs {
		accumulated += out.Amount
	}

	return accumulated
}
