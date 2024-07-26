package encoding

import (
	"chainnet/pkg/kernel"

	"github.com/stretchr/testify/mock"
)

type MockEncoding struct {
	mock.Mock
}

func (m *MockEncoding) SerializeBlock(b kernel.Block) ([]byte, error) {
	args := m.Called(b)
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockEncoding) DeserializeBlock(data []byte) (*kernel.Block, error) {
	args := m.Called(data)
	return args.Get(0).(*kernel.Block), args.Error(1)
}

func (m *MockEncoding) SerializeHeader(bh kernel.BlockHeader) ([]byte, error) {
	args := m.Called(bh)
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockEncoding) DeserializeHeader(data []byte) (*kernel.BlockHeader, error) {
	args := m.Called(data)
	return args.Get(0).(*kernel.BlockHeader), args.Error(1)
}

func (m *MockEncoding) SerializeTransaction(tx kernel.Transaction) ([]byte, error) {
	args := m.Called(tx)
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockEncoding) DeserializeTransaction(data []byte) (*kernel.Transaction, error) {
	args := m.Called(data)
	return args.Get(0).(*kernel.Transaction), args.Error(1)
}
