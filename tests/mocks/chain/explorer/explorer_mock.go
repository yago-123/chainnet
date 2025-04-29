package explorer

import (
	"github.com/stretchr/testify/mock"
	"github.com/yago-123/chainnet/pkg/kernel"
	"time"
)

type MockExplorer struct {
	mock.Mock
}

func (e *MockExplorer) GetLastBlock() (*kernel.Block, error) {
	args := e.Called()
	return args.Get(0).(*kernel.Block), args.Error(1)
}

func (e *MockExplorer) GetBlockByHash(hash []byte) (*kernel.Block, error) {
	args := e.Called(hash)
	return args.Get(0).(*kernel.Block), args.Error(1)
}

func (e *MockExplorer) GetHeaderByHeight(height uint) (*kernel.BlockHeader, error) {
	args := e.Called(height)
	return args.Get(0).(*kernel.BlockHeader), args.Error(1)
}

func (e *MockExplorer) GetLastHeader() (*kernel.BlockHeader, error) {
	args := e.Called()
	return args.Get(0).(*kernel.BlockHeader), args.Error(1)
}

func (e *MockExplorer) GetMiningTarget(height uint, difficultyAdjustmentInterval uint, expectedMiningInterval time.Duration) (uint, error) {
	args := e.Called(height, difficultyAdjustmentInterval, expectedMiningInterval)
	return args.Get(0).(uint), args.Error(1)
}

func (e *MockExplorer) GetAllHeaders() ([]*kernel.BlockHeader, error) {
	args := e.Called()
	return args.Get(0).([]*kernel.BlockHeader), args.Error(1)
}

func (e *MockExplorer) GetUnspentTransactions(address string) ([]*kernel.Transaction, error) {
	args := e.Called(address)
	return args.Get(0).([]*kernel.Transaction), args.Error(1)
}

func (e *MockExplorer) GetUnspentOutputs(address string, maxRetrievalNum int) ([]*kernel.UTXO, error) {
	args := e.Called(address, maxRetrievalNum)
	return args.Get(0).([]*kernel.UTXO), args.Error(1)
}

func (e *MockExplorer) GetAddressBalance(address string) (uint, error) {
	args := e.Called(address)
	return args.Get(0).(uint), args.Error(1)
}

func (e *MockExplorer) GetAmountSpendableOutputs(address string, amount uint) (uint, map[string][]uint, error) {
	args := e.Called(address, amount)
	return args.Get(0).(uint), args.Get(1).(map[string][]uint), args.Error(2)
}

func (e *MockExplorer) GetAllTransactions(address string, maxRetrievalNum int) ([]*kernel.Transaction, error) {
	args := e.Called(address, maxRetrievalNum)
	return args.Get(0).([]*kernel.Transaction), args.Error(1)
}

func (e *MockExplorer) GetUnspentTransactionsOutputs(address string) ([]kernel.TxOutput, error) {
	args := e.Called(address)
	return args.Get(0).([]kernel.TxOutput), args.Error(1)
}
