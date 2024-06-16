package consensus

import (
	"chainnet/pkg/block"
)

// Consensus is designed to allow more than one consensus algorithm to be implemented
type Consensus interface {
	ValidateBlock(block *block.Block) bool
	CalculateBlockHash(block *block.Block) (*block.Block, error)

	ValidateTx(tx *block.Transaction) bool
	CalculateTxHash(tx *block.Transaction) (*block.Transaction, error)
}
