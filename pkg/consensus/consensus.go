package consensus

import (
	"chainnet/pkg/kernel"
)

// Consensus is designed to allow more than one consensus algorithm to be implemented
type Consensus interface {
	ValidateTx(tx *kernel.Transaction) bool
	CalculateTxHash(tx *kernel.Transaction) ([]byte, error)

	ValidateBlock(b *kernel.Block) bool
	CalculateBlockHash(b *kernel.Block) ([]byte, uint, error)
}

// LightValidator represents a validator that does not require having the whole chain downloaded locally
// like for example the ones performed by wallets before sending transactions to the nodes and miners
type LightValidator interface {
	ValidateTx(tx *kernel.Transaction) bool
}

// HeavyValidator performs the same validations as LightValidator but also validates the previous
// information like the validity of the chain and transactions without funds. This validator is used
// by nodes and miners
type HeavyValidator interface {
	ValidateTx(tx *kernel.Transaction) bool
	ValidateBlock(b *kernel.Block) bool
}
