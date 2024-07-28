package hash

import (
	"bytes"
)

// FakeHashing adds "-hashed" to the payload provided so the hash can be predictable during tests
type FakeHashing struct {
}

func (mh *FakeHashing) Hash(payload []byte) ([]byte, error) {
	return append(payload, []byte("-hashed")...), nil
}

func (mh *FakeHashing) Verify(hash []byte, payload []byte) (bool, error) {
	h, _ := mh.Hash(payload)

	return bytes.Equal(hash, h), nil
}
