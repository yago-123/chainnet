package hash

import (
	"bytes"
)

// MockHashing adds "-hashed" to the payload provided so the hash can be predictable during tests
type MockHashing struct {
}

func (mh *MockHashing) Hash(payload []byte) []byte {
	return append(payload, []byte("-hashed")...)
}

func (mh *MockHashing) Verify(hash []byte, payload []byte) bool {
	return bytes.Equal(hash, mh.Hash(payload))
}
