package consensus

import (
	"chainnet/pkg/kernel"
)

// Consensus is designed to allow more than one consensus algorithm to be implemented
type Consensus interface {
	ValidateBlock(b *kernel.Block) bool
	CalculateBlockHash(b *kernel.Block) ([]byte, uint, error)

	ValidateTx(tx *kernel.Transaction) bool
	CalculateTxHash(tx *kernel.Transaction) ([]byte, error)
}
