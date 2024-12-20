package hash

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"hash"
	"sync"

	"golang.org/x/crypto/ripemd160"
)

const (
	SHA256 HasherType = iota
	RipeMD160
)

type HasherType uint

type Hashing interface {
	Hash(payload []byte) ([]byte, error)
	Verify(hashedPayload []byte, payload []byte) (bool, error)
}

// GetHasher represents a factory function that returns the hashing algorithm. This
// factory method is used in cases in which we need to use hashing in parallel hashing
// computations given that the algorithms are not thread safe
func GetHasher(i HasherType) Hashing {
	switch i {
	case SHA256:
		return NewHasher(sha256.New())
	case RipeMD160:
		return NewHasher(ripemd160.New())
	default:
		return NewVoidHasher()
	}
}

type Hash struct {
	h  hash.Hash
	mu sync.Mutex
}

func NewHasher(hashAlgo hash.Hash) *Hash {
	return &Hash{
		h: hashAlgo,
	}
}

func (hash *Hash) Hash(payload []byte) ([]byte, error) {
	if err := hashInputValidator(payload); err != nil {
		return []byte{}, err
	}

	hash.mu.Lock()
	defer hash.mu.Unlock()

	// reset the hasher state
	hash.h.Reset()

	_, err := hash.h.Write(payload)
	if err != nil {
		return []byte{}, err
	}

	return hash.h.Sum(nil), nil
}

func (hash *Hash) Verify(hashedPayload []byte, payload []byte) (bool, error) {
	if err := verifyInputValidator(hashedPayload, payload); err != nil {
		return false, err
	}

	h, err := hash.Hash(payload)
	if err != nil {
		return false, err
	}

	return bytes.Equal(hashedPayload, h), nil
}

func hashInputValidator(payload []byte) error {
	if len(payload) < 1 {
		return errors.New("payload is empty")
	}

	return nil
}

func verifyInputValidator(hash []byte, payload []byte) error {
	if len(hash) < 1 {
		return errors.New("hash is empty")
	}

	if len(payload) < 1 {
		return errors.New("payload is empty")
	}

	return nil
}
