package hash

import (
	"bytes"
	"crypto/sha256"
	"hash"
	"sync"
)

type Sha256 struct {
	sha hash.Hash
	mu  sync.Mutex
}

func NewSHA256() *Sha256 {
	return &Sha256{
		sha: sha256.New(),
	}
}

func (s *Sha256) Hash(payload []byte) ([]byte, error) {
	if err := hashInputValidator(payload); err != nil {
		return []byte{}, err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// reset the hasher state
	s.sha.Reset()

	_, err := s.sha.Write(payload)
	if err != nil {
		return []byte{}, err
	}

	return s.sha.Sum(nil), nil
}

func (s *Sha256) Verify(hash []byte, payload []byte) (bool, error) {
	if err := verifyInputValidator(hash, payload); err != nil {
		return false, err
	}

	h, err := s.Hash(payload)
	if err != nil {
		return false, err
	}

	return bytes.Equal(hash, h), nil
}
