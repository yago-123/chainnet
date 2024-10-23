package encoding

import (
	"github.com/yago-123/chainnet/pkg/kernel"
)

const (
	GobEncodingType   = "gob"
	ProtoEncodingType = "protobuf"
)

type Encoding interface {
	Type() string

	SerializeBlock(b kernel.Block) ([]byte, error)
	SerializeHeader(bh kernel.BlockHeader) ([]byte, error)
	SerializeHeaders(bhs []*kernel.BlockHeader) ([]byte, error)
	SerializeTransaction(tx kernel.Transaction) ([]byte, error)
	SerializeTransactions(txs []*kernel.Transaction) ([]byte, error)
	SerializeUTXO(utxo kernel.UTXO) ([]byte, error)
	SerializeUTXOs(utxos []*kernel.UTXO) ([]byte, error)

	DeserializeBlock(data []byte) (*kernel.Block, error)
	DeserializeHeader(data []byte) (*kernel.BlockHeader, error)
	DeserializeHeaders(data []byte) ([]*kernel.BlockHeader, error)
	DeserializeTransaction(data []byte) (*kernel.Transaction, error)
	DeserializeTransactions(data []byte) ([]*kernel.Transaction, error)
	DeserializeUTXO(data []byte) (*kernel.UTXO, error)
	DeserializeUTXOs(data []byte) ([]*kernel.UTXO, error)
}
