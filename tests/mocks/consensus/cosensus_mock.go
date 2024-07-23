package consensus

import (
	"chainnet/pkg/kernel"
	"context"

	"github.com/stretchr/testify/mock"
)

// MockConsensus is a mock implementation of Consensus interface for testing purposes
type MockConsensus struct {
	mock.Mock
}

func (m *MockConsensus) ValidateBlock(b *kernel.Block) bool {
	args := m.Called(b)
	return args.Bool(0)
}

func (m *MockConsensus) CalculateBlockHash(_ context.Context, b *kernel.Block) ([]byte, uint, error) {
	args := m.Called(b)
	return args.Get(0).([]byte), args.Get(1).(uint), args.Error(2)
}

func (m *MockConsensus) ValidateTx(tx *kernel.Transaction) bool {
	args := m.Called(tx)
	return args.Bool(0)
}

func (m *MockConsensus) CalculateTxHash(tx *kernel.Transaction) ([]byte, error) {
	args := m.Called(tx)
	return args.Get(0).([]byte), args.Error(1)
}
