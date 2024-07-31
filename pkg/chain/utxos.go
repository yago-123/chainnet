package blockchain

import (
	"chainnet/pkg/kernel"
	"fmt"
	"sync"
)

const UTXOSObserverID = "utxos"

type UTXOSet struct {
	mu    sync.Mutex
	utxos map[string]kernel.UnspentOutput
}

func NewUTXOSet() *UTXOSet {
	return &UTXOSet{
		mu:    sync.Mutex{},
		utxos: make(map[string]kernel.UnspentOutput),
	}
}

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

			_, ok := u.utxos[string(input.Txid)]
			if !ok {
				// if the utxo is not found, return error (impossible scenario in theory)
				return fmt.Errorf("transaction %s not found in the UTXO set", tx.ID)
			}

			// delete the utxo from the set
			delete(u.utxos, string(tx.ID))
		}

		// add new outputs to the set
		for index, output := range tx.Vout {
			utxo := kernel.UnspentOutput{
				TxID:   tx.ID,
				OutIdx: uint(index),
				Output: output,
			}

			// store utxo in the set
			u.utxos[string(tx.ID)] = utxo
		}
	}

	return nil
}

// ID returns the observer id
func (u *UTXOSet) ID() string {
	return UTXOSObserverID
}

// OnBlockAddition is called when a new block is added to the blockchain via the observer pattern
func (u *UTXOSet) OnBlockAddition(_ *kernel.Block) {

}
