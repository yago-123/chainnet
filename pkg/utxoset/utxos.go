package utxoset

import (
	"fmt"
	"sync"
	"unsafe"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/yago-123/chainnet/pkg/monitor"

	"github.com/sirupsen/logrus"
	"github.com/yago-123/chainnet/config"

	"github.com/yago-123/chainnet/pkg/kernel"
)

const UTXOSObserverID = "utxos-observer"

type UTXOSet struct {
	mu    sync.Mutex
	utxos map[string]kernel.UTXO

	logger *logrus.Logger
	cfg    *config.Config
}

func NewUTXOSet(cfg *config.Config) *UTXOSet {
	return &UTXOSet{
		mu:     sync.Mutex{},
		utxos:  make(map[string]kernel.UTXO),
		logger: cfg.Logger,
		cfg:    cfg,
	}
}

// AddBlock invalidates the new inputs of the block and adds the new outputs to the UTXO set
func (u *UTXOSet) AddBlock(block *kernel.Block) error {
	u.mu.Lock()
	defer u.mu.Unlock()

	for _, tx := range block.Transactions {
		// invalidate inputs used in the block
		for _, input := range tx.Vin {
			// skip Coinbase transactions
			if tx.IsCoinbase() {
				continue
			}

			_, ok := u.utxos[input.UniqueTxoKey()]
			if !ok {
				// if the utxo is not found, return error (impossible scenario in theory)
				return fmt.Errorf("transaction %s not found in the UTXO set", tx.ID)
			}

			// delete the utxo from the set
			delete(u.utxos, input.UniqueTxoKey())
		}

		// add new outputs to the set
		for index, output := range tx.Vout {
			utxo := kernel.UTXO{
				TxID:   tx.ID,
				OutIdx: uint(index),
				Output: output,
			}

			// store utxo in the set
			u.utxos[utxo.UniqueKey()] = utxo
		}
	}

	return nil
}

// RetrieveInputsBalance from the inputs provided
func (u *UTXOSet) RetrieveInputsBalance(inputs []kernel.TxInput) (uint, error) {
	u.mu.Lock()
	defer u.mu.Unlock()

	balance := uint(0)
	for _, input := range inputs {
		utxo, ok := u.utxos[input.UniqueTxoKey()]
		if !ok {
			return 0, fmt.Errorf("input %s not found in the UTXO set", input.UniqueTxoKey())
		}

		balance += utxo.Output.Amount
	}

	return balance, nil
}

// ID returns the observer id
func (u *UTXOSet) ID() string {
	return UTXOSObserverID
}

// OnBlockAddition is called when a new block is added to the blockchain via the observer pattern
func (u *UTXOSet) OnBlockAddition(block *kernel.Block) {
	err := u.AddBlock(block)
	if err != nil {
		u.logger.Errorf("error adding block to UTXO set: %s", err)
		return
	}
}

// OnTxAddition is called when a new tx is added to the mempool via the observer pattern
func (u *UTXOSet) OnTxAddition(_ *kernel.Transaction) {
	// do nothing
}

// RegisterMetrics registers the UTXO set metrics to the prometheus registry
func (u *UTXOSet) RegisterMetrics(register *prometheus.Registry) {
	monitor.NewMetric(register, monitor.Gauge, "utxo_set_num_outputs", "Number of outputs in the UTXO set",
		func() float64 {
			return float64(len(u.utxos))
		},
	)

	monitor.NewMetric(register, monitor.Gauge, "utxo_set_storage_size", "Size of the UTXO set in bytes",
		func() float64 {
			storage := uint(0)
			for _, utxo := range u.utxos {
				storage += utxo.Output.Size() + uint(len(utxo.TxID)) + uint(unsafe.Sizeof(uint(0)))
			}

			return float64(storage)
		},
	)

	monitor.NewMetric(register, monitor.Gauge, "utxo_set_output_balance", "A gauge containing the total balance of the UTXO set",
		func() float64 {
			totalBalance := uint(0)
			for _, utxo := range u.utxos {
				totalBalance += utxo.Amount()
			}

			return float64(totalBalance)
		},
	)
}
