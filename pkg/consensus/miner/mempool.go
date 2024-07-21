package miner

import (
	"chainnet/pkg/kernel"
	"sort"
)

type txFeePair struct {
	Transaction *kernel.Transaction
	Fee         uint
}

type MemPool []txFeePair

func NewMemPool() MemPool {
	return MemPool{}
}

func (m MemPool) Len() int           { return len(m) }
func (m MemPool) Swap(i, j int)      { m[i], m[j] = m[j], m[i] }
func (m MemPool) Less(i, j int) bool { return m[i].Fee > m[j].Fee }

// AppendTransaction adds a transaction to the MemPool sorting by highest transaction fee first
func (m *MemPool) AppendTransaction(tx *kernel.Transaction, fee uint) {
	*m = append(*m, txFeePair{Transaction: tx, Fee: fee})
	sort.Sort(m) // Ensure the MemPool is sorted after adding
}

// RetrieveTransactions retrieves the transactions from the MemPool with the highest fee
func (m *MemPool) RetrieveTransactions(maxNumberTxs uint) ([]*kernel.Transaction, uint) {
	txs := []*kernel.Transaction{}

	if uint(m.Len()) < maxNumberTxs {
		maxNumberTxs = uint(m.Len())
	}

	totalFee := uint(0)
	for _ = range maxNumberTxs {
		tx, fee := m.pop()

		txs = append(txs, tx)
		totalFee += fee
	}

	return txs, totalFee
}

// pop removes the highest fee transaction from the MemPool
func (m *MemPool) pop() (*kernel.Transaction, uint) {
	if len(*m) == 0 {
		return nil, 0
	}
	highestFeeTx := (*m)[0]
	*m = (*m)[1:]
	return highestFeeTx.Transaction, highestFeeTx.Fee
}
