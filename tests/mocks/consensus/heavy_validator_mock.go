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

func (m *MockHeavyValidator) ValidateTx(_ *kernel.Transaction) error {
	return nil
}

func (m *MockHeavyValidator) ValidateHeader(_ *kernel.BlockHeader) error {
	return nil
}

func (m *MockHeavyValidator) ValidateBlock(_ *kernel.Block) error {
	return nil
}
