package consensus

import (
	"chainnet/pkg/block"
)

// Consensus is designed to allow more than one consensus algorithm to be implemented
type Consensus interface {
	ValidateBlock(b *block.Block) bool
	CalculateBlockHash(b *block.Block) ([]byte, uint, error)

	ValidateTx(tx *block.Transaction) bool
	CalculateTxHash(tx *block.Transaction) ([]byte, error)
}
