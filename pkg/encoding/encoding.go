package encoding

import (
	"chainnet/pkg/block"
)

type Encoding interface {
	SerializeBlock(b block.Block) ([]byte, error)
	DeserializeBlock(data []byte) (*block.Block, error)

	SerializeTransaction(tx block.Transaction) ([]byte, error)
	DeserializeTransaction(data []byte) (*block.Transaction, error)
}
