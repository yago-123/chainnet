package blockchain

import "chainnet/pkg/kernel"

const UTXOSObserverId = "utxos"

type UTXOS struct {
}

func NewUTXOS() *UTXOS {
	return &UTXOS{}
}

// Id returns the observer id
func (u *UTXOS) Id() string {
	return UTXOSObserverId
}

// OnBlockAddition is called when a new block is added to the blockchain via the observer pattern
func (u *UTXOS) OnBlockAddition(block *kernel.Block) {

}
