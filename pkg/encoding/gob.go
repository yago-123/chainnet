package encoding

import (
	"bytes"
	"encoding/gob"
	"fmt"

	"github.com/yago-123/chainnet/pkg/kernel"
)

type GobEncoder struct {
}

func NewGobEncoder() *GobEncoder {
	return &GobEncoder{}
}

func (gobenc *GobEncoder) Type() string {
	return GobEncodingType
}

func (gobenc *GobEncoder) serialize(data interface{}) ([]byte, error) {
	var result bytes.Buffer
	encoder := gob.NewEncoder(&result)
	if err := encoder.Encode(data); err != nil {
		return nil, fmt.Errorf("error serializing data: %w", err)
	}
	return result.Bytes(), nil
}

func (gobenc *GobEncoder) deserialize(data []byte, out interface{}) error {
	decoder := gob.NewDecoder(bytes.NewReader(data))
	if err := decoder.Decode(out); err != nil {
		return fmt.Errorf("error deserializing data: %w", err)
	}
	return nil
}

func (gobenc *GobEncoder) SerializeBlock(b kernel.Block) ([]byte, error) {
	return gobenc.serialize(b)
}

func (gobenc *GobEncoder) DeserializeBlock(data []byte) (*kernel.Block, error) {
	var b kernel.Block
	if err := gobenc.deserialize(data, &b); err != nil {
		return nil, fmt.Errorf("error deserializing block: %w", err)
	}
	return &b, nil
}

func (gobenc *GobEncoder) SerializeHeader(bh kernel.BlockHeader) ([]byte, error) {
	return gobenc.serialize(bh)
}

func (gobenc *GobEncoder) DeserializeHeader(data []byte) (*kernel.BlockHeader, error) {
	var bh kernel.BlockHeader
	if err := gobenc.deserialize(data, &bh); err != nil {
		return nil, fmt.Errorf("error deserializing block header: %w", err)
	}
	return &bh, nil
}

func (gobenc *GobEncoder) SerializeHeaders(headers []*kernel.BlockHeader) ([]byte, error) {
	return gobenc.serialize(headers)
}

func (gobenc *GobEncoder) DeserializeHeaders(data []byte) ([]*kernel.BlockHeader, error) {
	var headers []*kernel.BlockHeader
	if err := gobenc.deserialize(data, &headers); err != nil {
		return nil, fmt.Errorf("error deserializing block headers: %w", err)
	}
	return headers, nil
}

func (gobenc *GobEncoder) SerializeTransaction(tx kernel.Transaction) ([]byte, error) {
	return gobenc.serialize(tx)
}

func (gobenc *GobEncoder) DeserializeTransaction(data []byte) (*kernel.Transaction, error) {
	var tx kernel.Transaction
	if err := gobenc.deserialize(data, &tx); err != nil {
		return nil, fmt.Errorf("error deserializing transaction: %w", err)
	}
	return &tx, nil
}

func (gobenc *GobEncoder) SerializeTransactions(txs []*kernel.Transaction) ([]byte, error) {
	return gobenc.serialize(txs)
}

func (gobenc *GobEncoder) DeserializeTransactions(data []byte) ([]*kernel.Transaction, error) {
	var txs []*kernel.Transaction
	if err := gobenc.deserialize(data, &txs); err != nil {
		return nil, fmt.Errorf("error deserializing transactions: %w", err)
	}
	return txs, nil
}

func (gobenc *GobEncoder) SerializeUTXO(utxo kernel.UTXO) ([]byte, error) {
	return gobenc.serialize(utxo)
}

func (gobenc *GobEncoder) DeserializeUTXO(data []byte) (*kernel.UTXO, error) {
	var utxo kernel.UTXO
	if err := gobenc.deserialize(data, &utxo); err != nil {
		return nil, fmt.Errorf("error deserializing UTXO: %w", err)
	}
	return &utxo, nil
}

func (gobenc *GobEncoder) SerializeUTXOs(utxos []*kernel.UTXO) ([]byte, error) {
	return gobenc.serialize(utxos)
}

func (gobenc *GobEncoder) DeserializeUTXOs(data []byte) ([]*kernel.UTXO, error) {
	var utxos []*kernel.UTXO
	if err := gobenc.deserialize(data, &utxos); err != nil {
		return nil, fmt.Errorf("error deserializing UTXOs: %w", err)
	}
	return utxos, nil
}

func (gobenc *GobEncoder) SerializeBool(b bool) ([]byte, error) {
	return gobenc.serialize(b)
}

func (gobenc *GobEncoder) DeserializeBool(data []byte) (bool, error) {
	var b bool
	if err := gobenc.deserialize(data, &b); err != nil {
		return false, fmt.Errorf("error deserializing bool: %w", err)
	}

	return b, nil
}
