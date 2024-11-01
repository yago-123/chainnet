package mempool

import (
	"sort"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/yago-123/chainnet/pkg/monitor"

	"errors"

	"github.com/yago-123/chainnet/pkg/kernel"
)

const MemPoolObserverID = "mempool-observer"

var ErrMemPoolFull = errors.New("mempool does not have enough space")

type TxFeePair struct {
	Transaction *kernel.Transaction
	Fee         uint
}

type MemPool struct {
	// pairs is a slice of transactions and their corresponding fees
	pairs []TxFeePair
	// txIDs is a map containing transaction ID as key and transaction as value
	txIDs map[string]*kernel.Transaction
	// inputSet is used to keep track of the inputs that are being spent in the mempool. This is useful for removing
	// transactions that are going to be invalid after a block addition. The key is the STXO key and the value is
	// the transaction ID that is spending it
	inputSet map[string][]string
	// maxNumberTxs is the maximum number of transactions the mempool can hold
	maxNumberTxs uint

	mu sync.Mutex
}

func NewMemPool(maxNumberTxs uint) *MemPool {
	return &MemPool{
		pairs:        make([]TxFeePair, 0, maxNumberTxs),
		txIDs:        make(map[string]*kernel.Transaction),
		inputSet:     make(map[string][]string),
		maxNumberTxs: maxNumberTxs,
	}
}

func (m *MemPool) Len() int           { return len(m.pairs) }
func (m *MemPool) Swap(i, j int)      { m.pairs[i], m.pairs[j] = m.pairs[j], m.pairs[i] }
func (m *MemPool) Less(i, j int) bool { return m.pairs[i].Fee > m.pairs[j].Fee }

// AppendTransaction adds a transaction to the MemPool sorting by highest transaction fee first
func (m *MemPool) AppendTransaction(tx *kernel.Transaction, fee uint) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if uint(len(m.pairs)) >= m.maxNumberTxs {
		return ErrMemPoolFull
	}

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

	m.txIDs[string(tx.ID)] = tx

	// ensure MemPool is sorted after adding (may be faster ways, but this is fine for now)
	sort.Sort(m)

	return nil
}

// ContainsTx checks if the MemPool contains a transaction with the given txID
func (m *MemPool) ContainsTx(txID string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	_, ok := m.txIDs[txID]
	return ok
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
			if txIDs, ok := m.inputSet[txInput.UniqueTxoKey()]; ok {
				// add the all txs that are spending the input to the removeTx map
				for _, txID := range txIDs {
					removeTx[txID] = true
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

	// remove the txs from the txIDs map
	for k := range removeTx {
		delete(m.txIDs, k)
	}
}

// OnTxAddition is called when a new tx is added to the mempool via the observer pattern
func (m *MemPool) OnTxAddition(_ *kernel.Transaction) {
	// do nothing
}

// RegisterMetrics registers the UTXO set metrics to the prometheus registry
func (m *MemPool) RegisterMetrics(register *prometheus.Registry) {
	monitor.NewMetric(register, monitor.Gauge, "mempool_size", "A gauge containing the number of transactions in the mempool",
		func() float64 {
			m.mu.Lock()
			defer m.mu.Unlock()

			return float64(len(m.pairs))
		},
	)

	monitor.NewMetric(register, monitor.Gauge, "mempool_total_fee", "A gauge containing the total fee of the transactions in the mempool",
		func() float64 {
			m.mu.Lock()
			defer m.mu.Unlock()

			totalFee := uint(0)
			for _, pair := range m.pairs {
				totalFee += pair.Fee
			}
			return float64(totalFee)
		},
	)

	monitor.NewMetric(register, monitor.Gauge, "mempool_max_size", "A gauge containing the maximum number of transactions the mempool can hold",
		func() float64 {
			return float64(m.maxNumberTxs)
		},
	)

	monitor.NewMetric(register, monitor.Gauge, "mempool_inputs_tracked", "A gauge containing the number of inputs being tracked in the mempool",
		func() float64 {
			m.mu.Lock()
			defer m.mu.Unlock()

			return float64(len(m.inputSet))
		},
	)

	// todo(): add histogram reflecting the distribution of fees in the mempool
	// todo(): add histogram reflecting the distribution of transaction sizes in the mempool
	// todo(): add histogram reflecting the distribution of transaction fees per byte in the mempool
	// todo(): add histogram reflecting the number of inputs per transaction in the mempool
	// todo(): add histogram reflecting the number of outputs per transaction in the mempool
}
