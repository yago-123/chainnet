package miner

import (
	"chainnet/pkg/kernel"
	"sort"
)

type txFeePair struct {
	Transaction *kernel.Transaction
	Fee         uint
}

type MemPool struct {
	pairs []txFeePair
}

func NewMemPool() MemPool {
	return MemPool{}
}

func (m MemPool) Len() int           { return len(m.pairs) }
func (m MemPool) Swap(i, j int)      { m.pairs[i], m.pairs[j] = m.pairs[j], m.pairs[i] }
func (m MemPool) Less(i, j int) bool { return m.pairs[i].Fee > m.pairs[j].Fee }

// AppendTransaction adds a transaction to the MemPool sorting by highest transaction fee first
func (m *MemPool) AppendTransaction(tx *kernel.Transaction, fee uint) {
	if tx == nil {
		return
	}

	// todo(): handle somehow the case in which multiple transactions contain multiple inputs from the same address
	// todo(): generating double spending transactions when mining a block wasting the miner resources after the block
	// todo(): is mined (block will be discarded by validators)

	m.pairs = append(m.pairs, txFeePair{Transaction: tx, Fee: fee})
	// ensure MemPool is sorted after adding (better ways of doing this really)
	sort.Sort(m)
}

// RetrieveTransactions retrieves the transactions from the MemPool with the highest fee
func (m *MemPool) RetrieveTransactions(maxNumberTxs uint) ([]*kernel.Transaction, uint) {
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
