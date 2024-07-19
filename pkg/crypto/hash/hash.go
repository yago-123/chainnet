package hash

import "errors"

type Hashing interface {
	Hash(payload []byte) ([]byte, error)
	Verify(hash []byte, payload []byte) (bool, error)
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
