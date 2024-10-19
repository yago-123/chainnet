package mempool

import (
	"github.com/yago-123/chainnet/pkg/kernel"
	"sort"
	"sync"
)

const MemPoolObserverID = "mempool-observer"

type TxFeePair struct {
	Transaction *kernel.Transaction
	Fee         uint
}

type MemPool struct {
	pairs []TxFeePair
	mu    sync.Mutex
}

func NewMemPool() *MemPool {
	return &MemPool{}
}

func (m *MemPool) Len() int           { return len(m.pairs) }
func (m *MemPool) Swap(i, j int)      { m.pairs[i], m.pairs[j] = m.pairs[j], m.pairs[i] }
func (m *MemPool) Less(i, j int) bool { return m.pairs[i].Fee > m.pairs[j].Fee }

// AppendTransaction adds a transaction to the MemPool sorting by highest transaction fee first
func (m *MemPool) AppendTransaction(tx *kernel.Transaction, fee uint) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if tx == nil {
		return
	}

	// todo(): handle somehow the case in which multiple transactions contain multiple inputs from the same address
	// todo(): generating double spending transactions when mining a block wasting the miner resources after the block
	// todo(): is mined (block will be discarded by validators)

	m.pairs = append(m.pairs, TxFeePair{Transaction: tx, Fee: fee})
	// ensure MemPool is sorted after adding (better ways of doing this really)
	sort.Sort(m)
}

// RetrieveTransactions retrieves the transactions from the MemPool with the highest fee
func (m *MemPool) RetrieveTransactions(maxNumberTxs uint) ([]*kernel.Transaction, uint) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if uint(m.Len()) < maxNumberTxs {
		maxNumberTxs = uint(m.Len())
	}

	// select the transactions
	tmpPairs := m.pairs[:maxNumberTxs]
	m.pairs = m.pairs[maxNumberTxs:]

	// calculate the total fee and return the transactions
	totalFee := uint(0)
	txs := []*kernel.Transaction{}
	for i := range tmpPairs {
		totalFee += tmpPairs[i].Fee
		txs = append(txs, tmpPairs[i].Transaction)
	}

	return txs, totalFee
}

// Id returns the observer id
func (m *MemPool) ID() string {
	return MemPoolObserverID
}

// OnBlockAddition is called when a new block is added to the blockchain via the observer pattern
func (m *MemPool) OnBlockAddition(_ *kernel.Block) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// todo(): we really need to do anything?
}
