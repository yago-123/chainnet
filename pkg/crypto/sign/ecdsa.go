package sign

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
)

type ECDSASigner struct {
}

func NewECDSASignature() *ECDSASigner {
	return &ECDSASigner{}
}

func (ecdsaSign *ECDSASigner) NewKeyPair() ([]byte, []byte, error) {
	curve := elliptic.P256()
	private, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		return []byte{}, []byte{}, err
	}

	pubKey := append(private.PublicKey.X.Bytes(), private.PublicKey.Y.Bytes()...)

	// todo() check that we can convert back to big.Int
	return private.D.Bytes(), pubKey, nil
}

func (ecdsaSign *ECDSASigner) Sign([]byte) ([]byte, error) {
	return []byte{}, nil
}

func (ecdsaSign *ECDSASigner) Verify([]byte, []byte) error {
	return nil
}