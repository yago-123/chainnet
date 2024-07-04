package sign

import "errors"

type Signature interface {
	NewKeyPair() ([]byte, []byte, error)
	// todo() should we add the transaction object here directly instead of payload?
	Sign(payload []byte, privKey []byte) ([]byte, error)
	Verify(signature []byte, payload []byte, pubKey []byte) (bool, error)
}

// signInputValidator controls the input validation for the sign method
func signInputValidator(payload []byte, privKey []byte) error {
	if len(payload) < 1 {
		return errors.New("payload is empty")
	}

	if len(privKey) < 1 {
		return errors.New("private key is empty")
	}

	return nil
}

// verifyInputValidator controls the input validation for the verify method
func verifyInputValidator(signature []byte, payload []byte, pubKey []byte) error {
	if len(signature) < 1 {
		return errors.New("signature is empty")
	}

	if len(payload) < 1 {
		return errors.New("payload is empty")
	}

	if len(pubKey) < 1 {
		return errors.New("public key is empty")
	}

	return nil
}
