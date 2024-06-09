package consensus

import (
	"chainnet/pkg/block"
)

type Consensus interface {
	Validate(block *block.Block) bool
	Calculate(block *block.Block) ([]byte, uint)
}
