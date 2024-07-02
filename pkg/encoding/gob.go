package encoding

import (
	"bytes"
	"chainnet/pkg/kernel"
	"encoding/gob"
	"fmt"
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
		return []byte{}, fmt.Errorf("error serializing block %s: %w", string(b.Hash), err)
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
