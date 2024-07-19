package hash

import (
	"bytes"
)

// MockHashing adds "-hashed" to the payload provided so the hash can be predictable during tests
type MockHashing struct {
}

func (mh *MockHashing) Hash(payload []byte) ([]byte, error) {
	return append(payload, []byte("-hashed")...), nil
}

func (mh *MockHashing) Verify(hash []byte, payload []byte) (bool, error) {
	h, _ := mh.Hash(payload)

	return bytes.Equal(hash, h), nil
}
