package encoding

import (
	"bytes"
	"chainnet/pkg/kernel"
	"encoding/gob"
	"github.com/sirupsen/logrus"
)

type GobEncoder struct {
	logger *logrus.Logger
}

func NewGobEncoder(logger *logrus.Logger) *GobEncoder {
	return &GobEncoder{logger}
}

func (gobenc *GobEncoder) SerializeBlock(b kernel.Block) ([]byte, error) {
	var result bytes.Buffer

	encoder := gob.NewEncoder(&result)
	err := encoder.Encode(b)
	if err != nil {
		return []byte{}, err
	}

	return result.Bytes(), nil
}

func (gobenc *GobEncoder) DeserializeBlock(data []byte) (*kernel.Block, error) {
	var b kernel.Block

	decoder := gob.NewDecoder(bytes.NewReader(data))
	err := decoder.Decode(&b)
	if err != nil {
		return &kernel.Block{}, err
	}

	return &b, nil
}

func (gobenc *GobEncoder) SerializeTransaction(tx kernel.Transaction) ([]byte, error) {
	var result bytes.Buffer

	enc := gob.NewEncoder(&result)
	err := enc.Encode(tx)
	if err != nil {
		return []byte{}, err
	}

	return result.Bytes(), nil
}

func (gobenc *GobEncoder) DeserializeTransaction(data []byte) (*kernel.Transaction, error) {
	var tx *kernel.Transaction

	decoder := gob.NewDecoder(bytes.NewReader(data))
	err := decoder.Decode(&tx)
	if err != nil {
		return &kernel.Transaction{}, err
	}

	return tx, nil
}
