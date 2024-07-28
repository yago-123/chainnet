package hash

import (
	"github.com/stretchr/testify/mock"
)

// MockHashing adds "-hashed" to the payload provided so the hash can be predictable during tests
type MockHashing struct {
	mock.Mock
}

func (mh *MockHashing) Hash(payload []byte) ([]byte, error) {
	args := mh.Called(payload)
	return args.Get(0).([]byte), args.Error(1)
}

func (mh *MockHashing) Verify(hash []byte, payload []byte) (bool, error) {
	args := mh.Called(hash, payload)
	return args.Bool(0), args.Error(1)
}
