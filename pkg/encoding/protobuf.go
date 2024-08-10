package encoding

import "chainnet/pkg/kernel"

type Protobuf struct {
}

func NewProtobufEncoder() *Protobuf {
	return &Protobuf{}
}

func (proto *Protobuf) SerializeBlock(b kernel.Block) ([]byte, error) {
	return []byte{}, nil
}

func (proto *Protobuf) SerializeHeader(bh kernel.BlockHeader) ([]byte, error) {
	return []byte{}, nil
}

func (proto *Protobuf) SerializeTransaction(tx kernel.Transaction) ([]byte, error) {
	return []byte{}, nil
}

func (proto *Protobuf) DeserializeBlock(data []byte) (*kernel.Block, error) {
	return &kernel.Block{}, nil
}

func (proto *Protobuf) DeserializeHeader(data []byte) (*kernel.BlockHeader, error) {
	return &kernel.BlockHeader{}, nil
}

func (proto *Protobuf) DeserializeTransaction(data []byte) (*kernel.Transaction, error) {
	return &kernel.Transaction{}, nil
}
