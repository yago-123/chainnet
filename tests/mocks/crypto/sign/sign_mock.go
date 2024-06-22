package sign

import "github.com/stretchr/testify/mock"

type MockSign struct {
	mock.Mock
}

func (m *MockSign) NewKeyPair() ([]byte, []byte, error) {
	args := m.Called()
	return args.Get(0).([]byte), args.Get(1).([]byte), args.Error(2)
}

func (m *MockSign) Sign(data []byte) ([]byte, error) {
	args := m.Called(data)
	return args.Get(0).([]byte), args.Error(1)
}
