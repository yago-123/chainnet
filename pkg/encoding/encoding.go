package encoding

import (
	"chainnet/pkg/kernel"
)

type Encoding interface {
	SerializeBlock(b kernel.Block) ([]byte, error)
	DeserializeBlock(data []byte) (*kernel.Block, error)

	SerializeTransaction(tx kernel.Transaction) ([]byte, error)
	DeserializeTransaction(data []byte) (*kernel.Transaction, error)
}
