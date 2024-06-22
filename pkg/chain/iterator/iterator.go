package iterator

import (
	"chainnet/pkg/block"
)

type Iterator interface {
	Initialize(reference []byte) error
	GetNextBlock() (*block.Block, error)
	HasNext() bool
}
