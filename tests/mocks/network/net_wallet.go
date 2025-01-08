package network

import (
	"context"
	"github.com/stretchr/testify/mock"
	"github.com/yago-123/chainnet/pkg/kernel"
)

type MockWalletNetwork struct {
	mock.Mock
}

func (m *MockWalletNetwork) GetWalletUTXOS(ctx context.Context, address []byte) ([]*kernel.UTXO, error) {
	args := m.Called(ctx, address)
	return args.Get(0).([]*kernel.UTXO), args.Error(1)
}

func (m *MockWalletNetwork) GetWalletTxs(ctx context.Context, address []byte) ([]*kernel.Transaction, error) {
	args := m.Called(ctx, address)
	return args.Get(0).([]*kernel.Transaction), args.Error(1)
}

func (m *MockWalletNetwork) SendTransaction(ctx context.Context, tx kernel.Transaction) error {
	args := m.Called(ctx, tx)
	return args.Error(0)
}
