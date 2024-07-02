package consensus

import (
	"chainnet/pkg/kernel"

	"github.com/stretchr/testify/mock"
)

type MockHeavyValidator struct {
	mock.Mock
}

func NewMockHeavyValidator() *MockHeavyValidator {
	return &MockHeavyValidator{}
}

func (m *MockHeavyValidator) ValidateTx(tx *kernel.Transaction) error {
	return nil
}

func (m *MockHeavyValidator) ValidateBlock(b *kernel.Block) error {
	return nil
}
