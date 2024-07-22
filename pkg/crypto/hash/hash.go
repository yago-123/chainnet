package hash

import "errors"

const (
	SHA256 HasherType = iota
	RipeMD160
)

type HasherType uint

type Hashing interface {
	Hash(payload []byte) ([]byte, error)
	Verify(hash []byte, payload []byte) (bool, error)
}

// GetHasher represents a factory function that returns the hashing algorithm. This
// factory method is used in cases in which we need to use hashing in parallel hashing
// computations given that the algorithms are not thread safe
func GetHasher(i HasherType) Hashing {
	switch i {
	case SHA256:
		return NewSHA256()
	case RipeMD160:
		return NewRipemd160()
	default:
		return NewVoidHasher()
	}
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
