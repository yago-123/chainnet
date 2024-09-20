package hash

import (
	"bytes"
	"hash"

	"golang.org/x/crypto/ripemd160" //nolint:staticcheck,gosec // need this lib as part of the specification
)

type Ripemd160 struct {
	ripe hash.Hash
}

func NewRipemd160() *Ripemd160 {
	return &Ripemd160{ripe: ripemd160.New()} //nolint:gosec // need this lib as part of the specification
}

func (r *Ripemd160) Hash(payload []byte) ([]byte, error) {
	if err := hashInputValidator(payload); err != nil {
		return []byte{}, err
	}

	return r.ripe.Sum(payload), nil
}

func (r *Ripemd160) Verify(hash []byte, payload []byte) (bool, error) {
	if err := verifyInputValidator(hash, payload); err != nil {
		return false, err
	}

	h, err := r.Hash(payload)
	if err != nil {
		return false, err
	}

	return bytes.Equal(hash, h), nil
}
