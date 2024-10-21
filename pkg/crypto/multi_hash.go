package crypto

import (
	"bytes"

	"github.com/yago-123/chainnet/pkg/crypto/hash"
)

type MultiHash struct {
	hashers []hash.Hashing
}

func NewMultiHash(hashers []hash.Hashing) *MultiHash {
	return &MultiHash{
		hashers: hashers,
	}
}

func (m *MultiHash) Hash(payload []byte) ([]byte, error) {
	var err error
	for _, hasher := range m.hashers {
		payload, err = hasher.Hash(payload)
		if err != nil {
			return []byte{}, err
		}
	}

	return payload, nil
}

func (m *MultiHash) Verify(hash []byte, payload []byte) (bool, error) {
	h, err := m.Hash(payload)
	if err != nil {
		return false, err
	}

	return bytes.Equal(hash, h), nil
}
