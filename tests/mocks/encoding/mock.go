package encoding

import (
	"chainnet/pkg/block"
	"github.com/stretchr/testify/mock"
)

type MockEncoding struct {
	mock.Mock
}

func (m *MockEncoding) SerializeBlock(b block.Block) ([]byte, error) {
	args := m.Called(b)
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockEncoding) DeserializeBlock(data []byte) (*block.Block, error) {
	args := m.Called(data)
	return args.Get(0).(*block.Block), args.Error(1)
}

func (m *MockEncoding) SerializeTransaction(tx block.Transaction) ([]byte, error) {
	args := m.Called(tx)
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockEncoding) DeserializeTransaction(data []byte) (*block.Transaction, error) {
	args := m.Called(data)
	return args.Get(0).(*block.Transaction), args.Error(1)
}
