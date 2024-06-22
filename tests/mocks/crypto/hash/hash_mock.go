package hash

import "github.com/stretchr/testify/mock"

type MockHashing struct {
	mock.Mock
}

func (mh *MockHashing) Hash(payload []byte) []byte {
	args := mh.Called(payload)
	return args.Get(0).([]byte)
}

func (mh *MockHashing) Verify(hash []byte, payload []byte) bool {
	args := mh.Called(hash, payload)
	return args.Bool(0)
}
