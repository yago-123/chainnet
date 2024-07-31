package blockchain

import (
	"chainnet/pkg/kernel"
	"fmt"
	"sync"
)

const UTXOSObserverID = "utxos"

type UTXOSet struct {
	mu    sync.Mutex
	utxos map[string]kernel.UTXO
}

func NewUTXOSet() *UTXOSet {
	return &UTXOSet{
		mu:    sync.Mutex{},
		utxos: make(map[string]kernel.UTXO),
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

// ID returns the observer id
func (u *UTXOSet) ID() string {
	return UTXOSObserverID
}

// OnBlockAddition is called when a new block is added to the blockchain via the observer pattern
func (u *UTXOSet) OnBlockAddition(block *kernel.Block) {
	err := u.AddBlock(block)
	if err != nil {
		// add logging
		return
	}
}
