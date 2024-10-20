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

func (gobenc *GobEncoder) SerializeBlock(b kernel.Block) ([]byte, error) {
	var result bytes.Buffer

	encoder := gob.NewEncoder(&result)
	err := encoder.Encode(b)
	if err != nil {
		return []byte{}, fmt.Errorf("error serializing block %x: %w", b.Hash, err)
	}

	return result.Bytes(), nil
}

func (gobenc *GobEncoder) DeserializeBlock(data []byte) (*kernel.Block, error) {
	var b kernel.Block

	decoder := gob.NewDecoder(bytes.NewReader(data))
	err := decoder.Decode(&b)
	if err != nil {
		return &kernel.Block{}, fmt.Errorf("error deserializing block: %w", err)
	}

	return &b, nil
}

func (gobenc *GobEncoder) SerializeHeader(bh kernel.BlockHeader) ([]byte, error) {
	var result bytes.Buffer

	encoder := gob.NewEncoder(&result)
	err := encoder.Encode(bh)
	if err != nil {
		return []byte{}, fmt.Errorf("error serializing block header: %w", err)
	}

	return result.Bytes(), nil
}

func (gobenc *GobEncoder) DeserializeHeader(data []byte) (*kernel.BlockHeader, error) {
	var bh kernel.BlockHeader

	decoder := gob.NewDecoder(bytes.NewReader(data))
	err := decoder.Decode(&bh)
	if err != nil {
		return &kernel.BlockHeader{}, fmt.Errorf("error deserializing block header: %w", err)
	}

	return &bh, nil
}

func (gobenc *GobEncoder) SerializeHeaders(headers []*kernel.BlockHeader) ([]byte, error) {
	var result bytes.Buffer

	encoder := gob.NewEncoder(&result)
	err := encoder.Encode(headers)
	if err != nil {
		return nil, fmt.Errorf("error serializing block headers: %w", err)
	}

	return result.Bytes(), nil
}

func (gobenc *GobEncoder) DeserializeHeaders(data []byte) ([]*kernel.BlockHeader, error) {
	var headers []*kernel.BlockHeader

	decoder := gob.NewDecoder(bytes.NewReader(data))
	err := decoder.Decode(&headers)
	if err != nil {
		return nil, fmt.Errorf("error deserializing block headers: %w", err)
	}

	return headers, nil
}

func (gobenc *GobEncoder) SerializeTransaction(tx kernel.Transaction) ([]byte, error) {
	var result bytes.Buffer

	enc := gob.NewEncoder(&result)
	err := enc.Encode(tx)
	if err != nil {
		return []byte{}, fmt.Errorf("error serializing transaction %s: %w", string(tx.ID), err)
	}

	return result.Bytes(), nil
}

func (gobenc *GobEncoder) DeserializeTransaction(data []byte) (*kernel.Transaction, error) {
	var tx *kernel.Transaction

	decoder := gob.NewDecoder(bytes.NewReader(data))
	err := decoder.Decode(&tx)
	if err != nil {
		return &kernel.Transaction{}, fmt.Errorf("error deserializing transaction: %w", err)
	}

	return tx, nil
}
