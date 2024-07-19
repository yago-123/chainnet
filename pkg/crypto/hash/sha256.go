package hash

import (
	"bytes"
	"crypto/sha256"
)

type Sha256 struct {
}

func NewSHA256() *Sha256 {
	return &Sha256{}
}

func (s *Sha256) Hash(payload []byte) ([]byte, error) {
	if err := hashInputValidator(payload); err != nil {
		return []byte{}, err
	}

	ret := sha256.Sum256(payload)
	return ret[:], nil
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
