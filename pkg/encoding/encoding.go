package encoding

import (
	"github.com/yago-123/chainnet/pkg/kernel"
)

type Encoding interface {
	SerializeBlock(b kernel.Block) ([]byte, error)
	SerializeHeader(bh kernel.BlockHeader) ([]byte, error)
	SerializeHeaders(bhs []*kernel.BlockHeader) ([]byte, error)
	SerializeTransaction(tx kernel.Transaction) ([]byte, error)

	DeserializeBlock(data []byte) (*kernel.Block, error)
	DeserializeHeader(data []byte) (*kernel.BlockHeader, error)
	DeserializeHeaders(data []byte) ([]*kernel.BlockHeader, error)
	DeserializeTransaction(data []byte) (*kernel.Transaction, error)
}
