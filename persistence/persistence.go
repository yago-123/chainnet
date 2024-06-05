package persistence

import (
	"chainnet/block"
)

type Persistence interface {
	SerializeBlock(b block.Block) []byte
}
