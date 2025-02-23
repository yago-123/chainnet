//nolint:errcheck // this is a mock
package sign

import (
	"bytes"

	"github.com/stretchr/testify/mock"
)

// MockSign adds "-signed" to the payload provided so the signature can be predictable during tests
type MockSign struct {
	mock.Mock
}

func (m *MockSign) NewKeyPair() ([]byte, []byte, error) {
	args := m.Called()
	return args.Get(0).([]byte), args.Get(1).([]byte), args.Error(2)
}

func (m *MockSign) Sign(payload []byte, _ []byte) ([]byte, error) {
	return append(payload, []byte("-signed")...), nil
}

func (m *MockSign) Verify(signature []byte, payload []byte, _ []byte) (bool, error) {
	return bytes.Equal(signature, append(payload, []byte("-signed")...)), nil
}
