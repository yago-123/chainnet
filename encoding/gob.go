package encoding

import (
	"bytes"
	"chainnet/block"
	"encoding/gob"
	"github.com/sirupsen/logrus"
)

type GobEncoder struct {
	logger *logrus.Logger
}

func NewGobEncoder(logger *logrus.Logger) *GobEncoder {
	return &GobEncoder{logger}
}

func (gobenc *GobEncoder) SerializeBlock(b block.Block) ([]byte, error) {
	var result bytes.Buffer

	encoder := gob.NewEncoder(&result)
	err := encoder.Encode(b)
	if err != nil {
		return []byte{}, err
	}

	return result.Bytes(), nil
}

func (gobenc *GobEncoder) DeserializeBlock(data []byte) (*block.Block, error) {
	var b block.Block

	decoder := gob.NewDecoder(bytes.NewReader(data))
	err := decoder.Decode(&b)
	if err != nil {
		return &block.Block{}, err
	}

	return &b, nil
}
