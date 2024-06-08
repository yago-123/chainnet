package consensus

import (
	"chainnet/block"
)

type Consensus interface {
	Validate(block *block.Block) bool
	Calculate(block *block.Block) ([]byte, uint)
}
