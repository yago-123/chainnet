//nolint:errcheck // this is a mock
package iterator

import (
	"github.com/yago-123/chainnet/pkg/kernel"

	"github.com/stretchr/testify/mock"
)

type MockIterator struct {
	mock.Mock
}

func (i *MockIterator) Initialize(reference []byte) error {
	args := i.Called(reference)
	return args.Error(0)
}

func (i *MockIterator) GetNextBlock() (*kernel.Block, error) {
	args := i.Called()
	return args.Get(0).(*kernel.Block), args.Error(1)
}

func (i *MockIterator) HasNext() bool {
	args := i.Called()
	return args.Bool(0)
}
