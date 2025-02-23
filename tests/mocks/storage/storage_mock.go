//nolint:errcheck // this is a mock
package storage

import (
	"github.com/yago-123/chainnet/pkg/kernel"

	"github.com/stretchr/testify/mock"
)

type MockStorage struct {
	mock.Mock
}

func (ms *MockStorage) PersistBlock(block kernel.Block) error {
	args := ms.Called(block)
	return args.Error(0)
}

func (ms *MockStorage) PersistHeader(_ []byte, _ kernel.BlockHeader) error {
	return nil
}

func (ms *MockStorage) GetLastBlock() (*kernel.Block, error) {
	args := ms.Called()
	return args.Get(0).(*kernel.Block), args.Error(1)
}

func (ms *MockStorage) GetLastHeader() (*kernel.BlockHeader, error) {
	args := ms.Called()
	return args.Get(0).(*kernel.BlockHeader), args.Error(1)
}

func (ms *MockStorage) GetLastBlockHash() ([]byte, error) {
	args := ms.Called()
	return args.Get(0).([]byte), args.Error(1)
}

func (ms *MockStorage) GetGenesisBlock() (*kernel.Block, error) {
	args := ms.Called()
	return args.Get(0).(*kernel.Block), args.Error(1)
}

func (ms *MockStorage) GetGenesisHeader() (*kernel.BlockHeader, error) {
	args := ms.Called()
	return args.Get(0).(*kernel.BlockHeader), args.Error(1)
}

func (ms *MockStorage) RetrieveBlockByHash(hash []byte) (*kernel.Block, error) {
	args := ms.Called(hash)
	return args.Get(0).(*kernel.Block), args.Error(1)
}

func (ms *MockStorage) RetrieveHeaderByHash(hash []byte) (*kernel.BlockHeader, error) {
	args := ms.Called(hash)
	return args.Get(0).(*kernel.BlockHeader), args.Error(1)
}

func (ms *MockStorage) Typ() string {
	return "mock"
}

func (ms *MockStorage) ID() string {
	return ms.Called().String(0)
}

func (ms *MockStorage) OnBlockAddition(block *kernel.Block) {
	ms.Called(block)
}

func (ms *MockStorage) OnTxAddition(tx *kernel.Transaction) {
	ms.Called(tx)
}

func (ms *MockStorage) Close() error {
	args := ms.Called()
	return args.Error(0)
}
