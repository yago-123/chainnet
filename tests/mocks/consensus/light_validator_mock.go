package consensus

import (
	"github.com/yago-123/chainnet/pkg/kernel"

	"github.com/stretchr/testify/mock"
)

type MockLightValidator struct {
	mock.Mock
}

func (m *MockLightValidator) ValidateTxLight(tx *kernel.Transaction) error {
	args := m.Called(tx)
	return args.Error(0)
}

func (m *MockLightValidator) ValidateHeader(bh *kernel.BlockHeader) error {
	args := m.Called(bh)
	return args.Error(0)
}
