package mempool

import (
	"github.com/yago-123/chainnet/pkg/chain/explorer"
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
	mu    sync.Mutex

	explorer *explorer.Explorer
}

func NewMemPool(explorer *explorer.Explorer) *MemPool {
	return &MemPool{
		explorer: explorer,
	}
}

func (m *MemPool) Len() int           { return len(m.pairs) }
func (m *MemPool) Swap(i, j int)      { m.pairs[i], m.pairs[j] = m.pairs[j], m.pairs[i] }
func (m *MemPool) Less(i, j int) bool { return m.pairs[i].Fee > m.pairs[j].Fee }

// AppendTransaction adds a transaction to the MemPool sorting by highest transaction fee first
func (m *MemPool) AppendTransaction(tx *kernel.Transaction) error {
	// retrieve balance from inputs
	inputBalance, err := m.explorer.CalculateBalanceFromInputs(tx.Vin)
	if err != nil {
		return err
	}

	// retrieve balance from outputs
	outputBalance := uint(0)
	for i := range tx.Vout {
		outputBalance += tx.Vout[i].Amount
	}

	// calculate fee of the transaction
	fee := inputBalance - outputBalance

	// lock mempool to make sure that no other transaction is added while we are adding this one
	m.mu.Lock()
	defer m.mu.Unlock()

	// todo(): make sure that there are no transactions with same tx.ID :)))

	// append the transaction to the mempool
	m.pairs = append(m.pairs, TxFeePair{Transaction: tx, Fee: fee})

	// ensure MemPool is sorted after adding (may be faster ways, but this is fine for now)
	sort.Sort(m)

	return nil
}

// RetrieveTransactions retrieves the transactions from the MemPool with the highest fee
func (m *MemPool) RetrieveTransactions(maxNumberTxs uint) ([]*kernel.Transaction, uint) {
	m.mu.Lock()
	defer m.mu.Unlock()

	totalFee := uint(0)
	txs := []*kernel.Transaction{}
	for i := range m.pairs {
		// make sure that the transactions retrieved do not contain other txs having same inputs. Otherwise the
		// miner will be mining blocks that will be discarded by the validator
		tmpTx := m.pairs[i].Transaction
		for _, tmpInput := range tmpTx.Vin {
			for _, txAdded := range txs {
				for _, inputAdded := range txAdded.Vin {
					if tmpInput.EqualInput(inputAdded) {
						continue
					}
				}
			}
		}

		// add the transaction to the list
		txs = append(txs, m.pairs[i].Transaction)
		totalFee += m.pairs[i].Fee

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
func (m *MemPool) OnBlockAddition(_ *kernel.Block) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// todo(): remove txs that contain inputs in the block that are spent in the block
}
