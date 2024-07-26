// storage.go
package storage

import (
	"chainnet/pkg/kernel"

	"github.com/stretchr/testify/mock"
)

type MockStorage struct {
	mock.Mock
}

func (ms *MockStorage) NumberOfBlocks() (uint, error) {
	args := ms.Called()
	return args.Get(0).(uint), args.Error(1)
}

func (ms *MockStorage) PersistBlock(block kernel.Block) error {
	args := ms.Called(block)
	return args.Error(0)
}

func (ms *MockStorage) PersistHeader(blockHash []byte, blockHeader kernel.BlockHeader) error {
	args := ms.Called(blockHash, blockHeader)
	return args.Error(0)
}

func (ms *MockStorage) GetLastBlock() (*kernel.Block, error) {
	args := ms.Called()
	return args.Get(0).(*kernel.Block), args.Error(1)
}

func (ms *MockStorage) GetGenesisBlock() (*kernel.Block, error) {
	args := ms.Called()
	return args.Get(0).(*kernel.Block), args.Error(1)
}

func (ms *MockStorage) RetrieveBlockByHash(hash []byte) (*kernel.Block, error) {
	args := ms.Called(hash)
	return args.Get(0).(*kernel.Block), args.Error(1)
}

func (ms *MockStorage) RetrieveHeaderByHash(hash []byte) (*kernel.BlockHeader, error) {
	args := ms.Called(hash)
	return args.Get(0).(*kernel.BlockHeader), args.Error(1)
}

func (ms *MockStorage) Close() error {
	args := ms.Called()
	return args.Error(0)
}
