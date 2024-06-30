package consensus

import (
	"chainnet/pkg/kernel"
	"github.com/stretchr/testify/mock"
)

type MockHeavyValidator struct {
	mock.Mock
}

func (m *MockHeavyValidator) ValidateTx(tx *kernel.Transaction) error {
	args := m.Called(tx)
	return args.Error(0)
}

func (m *MockHeavyValidator) ValidateBlock(b *kernel.Block) error {
	args := m.Called(b)
	return args.Error(0)
}
