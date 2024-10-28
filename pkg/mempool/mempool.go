package mempool

import (
	"sort"
	"sync"

	"github.com/yago-123/chainnet/pkg/kernel"
)

const MemPoolObserverID = "mempool-observer"

type TxFeePair struct {
	Transaction *kernel.Transaction
	Fee         uint
}

type MemPool struct {
	pairs []TxFeePair
	// inputSet is used to keep track of the inputs that are being spent in the mempool. This is useful for removing
	// transactions that are going to be invalid after a block addition. The key is the STXO key and the value is
	// the transaction ID that is spending it
	inputSet map[string][]string

	mu sync.Mutex
}

func NewMemPool() *MemPool {
	return &MemPool{
		pairs:    []TxFeePair{},
		inputSet: make(map[string][]string),
	}
}

func (m *MemPool) Len() int           { return len(m.pairs) }
func (m *MemPool) Swap(i, j int)      { m.pairs[i], m.pairs[j] = m.pairs[j], m.pairs[i] }
func (m *MemPool) Less(i, j int) bool { return m.pairs[i].Fee > m.pairs[j].Fee }

// AppendTransaction adds a transaction to the MemPool sorting by highest transaction fee first
func (m *MemPool) AppendTransaction(tx *kernel.Transaction, fee uint) error {
	// lock mempool to make sure that no other transaction is added while we are adding this one
	m.mu.Lock()
	defer m.mu.Unlock()

	// append the transaction to the mempool
	m.pairs = append(m.pairs, TxFeePair{Transaction: tx, Fee: fee})

	// append the inputs to inputSet to keep track of which inputs are being spent in which txs
	// this is useful for removing txs that are going to be invalid after a block addition
	// see OnBlockAddition function
	for _, v := range tx.Vin {
		if _, ok := m.inputSet[v.UniqueTxoKey()]; !ok {
			m.inputSet[v.UniqueTxoKey()] = []string{}
		}

		m.inputSet[v.UniqueTxoKey()] = append(m.inputSet[v.UniqueTxoKey()], string(tx.ID))
	}

	// ensure MemPool is sorted after adding (may be faster ways, but this is fine for now)
	sort.Sort(m)

	return nil
}

// RetrieveTransactions retrieves the transactions from the MemPool with the highest fee
func (m *MemPool) RetrieveTransactions(maxNumberTxs uint) ([]*kernel.Transaction, uint) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if maxNumberTxs == 0 {
		return []*kernel.Transaction{}, 0
	}

	totalFee := uint(0)
	txs := make([]*kernel.Transaction, 0, maxNumberTxs)
	retrievedInputs := map[string]bool{}

	for _, pair := range m.pairs {
		// make sure that the transactions retrieved do not contain other txs having same inputs. Otherwise the
		// miner will be mining blocks that will be discarded by the validator
		transaction := pair.Transaction
		hasConflictingInputs := false

		// check if the transaction has conflicting inputs
		for _, input := range transaction.Vin {
			// if input have already been added, mark transaction as having conflicting inputs
			if _, ok := retrievedInputs[input.UniqueTxoKey()]; ok {
				hasConflictingInputs = true
				break // exit the inner loop if a conflict is found
			}
		}

		// if there are conflicting inputs, skip adding this transaction
		if hasConflictingInputs {
			continue
		}

		// add the transaction to the list
		txs = append(txs, transaction)
		totalFee += pair.Fee

		// mark the inputs as used
		for _, input := range transaction.Vin {
			retrievedInputs[input.UniqueTxoKey()] = true
		}

		// stop looking for txs if already reached the goal
		if uint(len(txs)) == maxNumberTxs {
			break
		}
	}

	return txs, totalFee
}

// ID returns the observer id
func (m *MemPool) ID() string {
	return MemPoolObserverID
}

// OnBlockAddition is called when a new block is added to the blockchain via the observer pattern
func (m *MemPool) OnBlockAddition(block *kernel.Block) {
	m.mu.Lock()
	defer m.mu.Unlock()

	removeTx := map[string]bool{}
	for _, tx := range block.Transactions {
		// iterate over the inputs contained in the transaction
		for _, txInput := range tx.Vin {
			// if the input is in the inputSet, remove the txs that are spending it by adding them
			// into the map removeTx
			if txIds, ok := m.inputSet[txInput.UniqueTxoKey()]; ok {
				// add the all txs that are spending the input to the removeTx map
				for _, txId := range txIds {
					removeTx[txId] = true
				}
			}

			// remove the input from the inputSet
			delete(m.inputSet, txInput.UniqueTxoKey())
		}
	}

	// remove txs that contain inputs in the block that are spent in the block
	for i, tx := range m.pairs {
		if _, ok := removeTx[string(tx.Transaction.ID)]; ok {
			m.pairs = append(m.pairs[:i], m.pairs[i+1:]...)
		}
	}
}
