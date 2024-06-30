package consensus

import (
	"chainnet/pkg/kernel"
	"github.com/stretchr/testify/mock"
)

type MockLightValidator struct {
	mock.Mock
}

func (m *MockLightValidator) ValidateTxLight(tx *kernel.Transaction) error {
	args := m.Called(tx)
	return args.Error(0)
}
