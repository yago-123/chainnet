package consensus

import (
	"chainnet/pkg/block"
	"github.com/stretchr/testify/mock"
)

// MockConsensus is a mock implementation of Consensus interface for testing purposes
type MockConsensus struct {
	mock.Mock
}

func (m *MockConsensus) ValidateBlock(b *block.Block) bool {
	args := m.Called(b)
	return args.Bool(0)
}

func (m *MockConsensus) CalculateBlockHash(b *block.Block) ([]byte, uint, error) {
	args := m.Called(b)
	return args.Get(0).([]byte), args.Get(1).(uint), args.Error(2)
}

func (m *MockConsensus) ValidateTx(tx *block.Transaction) bool {
	args := m.Called(tx)
	return args.Bool(0)
}

func (m *MockConsensus) CalculateTxHash(tx *block.Transaction) ([]byte, error) {
	args := m.Called(tx)
	return args.Get(0).([]byte), args.Error(1)
}
