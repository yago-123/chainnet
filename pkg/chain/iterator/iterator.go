package iterator

import (
	"chainnet/pkg/kernel"
)

type Iterator interface {
	Initialize(reference []byte) error
	GetNextBlock() (*kernel.Block, error)
	HasNext() bool
}
