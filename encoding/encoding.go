package encoding

import (
	"chainnet/block"
)

type Encoding interface {
	SerializeBlock(b block.Block) ([]byte, error)
	DeserializeBlock(data []byte) (*block.Block, error)
}
