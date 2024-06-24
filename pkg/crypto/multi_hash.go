package crypto

import (
	"bytes"
	"chainnet/pkg/crypto/hash"
	"errors"
)

type MultiHash struct {
	hashers []hash.Hashing
}

func NewMultiHash(hashers []hash.Hashing) (*MultiHash, error) {
	if len(hashers) < 1 {
		return nil, errors.New("unable to start a multihasher with 0 or 1 hashers")
	}

	return &MultiHash{
		hashers: hashers,
	}, nil
}

func (m *MultiHash) Hash(payload []byte) []byte {
	for _, hasher := range m.hashers {
		payload = hasher.Hash(payload)
	}

	return payload
}

func (m *MultiHash) Verify(hash []byte, payload []byte) bool {
	return bytes.Equal(hash, m.Hash(payload))
}
