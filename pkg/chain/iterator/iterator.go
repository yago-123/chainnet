package iterator

import (
	"github.com/yago-123/chainnet/pkg/kernel"
)

type BlockIterator interface {
	Initialize(reference []byte) error
	GetNextBlock() (*kernel.Block, error)
	HasNext() bool
}

type HeaderIterator interface {
	Initialize(reference []byte) error
	GetNextHeader() (*kernel.BlockHeader, error)
	HasNext() bool
}
