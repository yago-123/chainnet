package encoding

import (
	"chainnet/pkg/block"
)

type Encoding interface {
	SerializeBlock(b block.Block) ([]byte, error)
	DeserializeBlock(data []byte) (*block.Block, error)
}
