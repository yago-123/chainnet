// storage.go
package storage

import (
	"chainnet/pkg/block"
	"github.com/stretchr/testify/mock"
)

// Storage interface remains the same
type Storage interface {
	NumberOfBlocks() (uint, error)
	PersistBlock(block block.Block) error
	GetLastBlock() (*block.Block, error)
	RetrieveBlockByHash(hash []byte) (*block.Block, error)
}

// MockStorage struct modified to embed testify/mock
type MockStorage struct {
	mock.Mock
}

// NewMockStorage initializes MockStorage
func NewMockStorage() *MockStorage {
	return &MockStorage{}
}

func (ms *MockStorage) NumberOfBlocks() (uint, error) {
	args := ms.Called()
	return args.Get(0).(uint), args.Error(1)
}

func (ms *MockStorage) PersistBlock(block block.Block) error {
	args := ms.Called(block)
	return args.Error(0)
}

func (ms *MockStorage) GetLastBlock() (*block.Block, error) {
	args := ms.Called()
	return args.Get(0).(*block.Block), args.Error(1)
}

func (ms *MockStorage) RetrieveBlockByHash(hash []byte) (*block.Block, error) {
	args := ms.Called(hash)
	return args.Get(0).(*block.Block), args.Error(1)
}
