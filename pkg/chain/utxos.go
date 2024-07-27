package blockchain

import "chainnet/pkg/kernel"

const UTXOSObserverID = "utxos"

type UTXOS struct {
}

func NewUTXOS() *UTXOS {
	return &UTXOS{}
}

// Id returns the observer id
func (u *UTXOS) ID() string {
	return UTXOSObserverID
}

// OnBlockAddition is called when a new block is added to the blockchain via the observer pattern
func (u *UTXOS) OnBlockAddition(_ *kernel.Block) {

}
