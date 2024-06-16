package consensus

import (
	"chainnet/pkg/block"
	"github.com/stretchr/testify/mock"
)

type MockConsensus struct {
	mock.Mock
}

func (m *MockConsensus) ValidateBlock(b *block.Block) bool {
	args := m.Called(b)
	return args.Bool(0)
}

func (m *MockConsensus) CalculateBlockHash(b *block.Block) (*block.Block, error) {
	args := m.Called(b)
	return args.Get(0).(*block.Block), args.Error(1)
}

func (m *MockConsensus) ValidateTx(tx *block.Transaction) bool {
	args := m.Called(tx)
	return args.Bool(0)
}

func (m *MockConsensus) CalculateTxHash(tx *block.Transaction) (*block.Transaction, error) {
	args := m.Called(tx)
	return args.Get(0).(*block.Transaction), args.Error(1)
}
