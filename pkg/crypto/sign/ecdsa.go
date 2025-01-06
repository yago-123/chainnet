package sign

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"math/big"

	util_crypto "github.com/yago-123/chainnet/pkg/util/crypto"
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

	return util_crypto.ConvertECDSAKeysToDERBytes(&private.PublicKey, private)
}

func (ecdsaSign *ECDSASigner) Sign(payload []byte, privKey []byte) ([]byte, error) {
	if err := signInputValidator(payload, privKey); err != nil {
		return []byte{}, err
	}

	privateKey, err := util_crypto.ConvertDERBytesToECDSAPriv(privKey)
	if err != nil {
		return []byte{}, err
	}

	r, s, err := ecdsa.Sign(rand.Reader, privateKey, payload)
	if err != nil {
		return []byte{}, err
	}

	// consolidate signature
	signature := r.Bytes()
	signature = append(signature, s.Bytes()...)

	return signature, nil
}

func (ecdsaSign *ECDSASigner) Verify(signature []byte, payload []byte, pubKey []byte) (bool, error) {
	if err := verifyInputValidator(signature, payload, pubKey); err != nil {
		return false, err
	}

	publicKey, err := util_crypto.ConvertDERBytesToECDSAPub(pubKey)
	if err != nil {
		return false, err
	}

	rLength := len(signature) / 2 //nolint:mnd  // we need to divide the signature in half
	rBytes := signature[:rLength]
	sBytes := signature[rLength:]

	r := new(big.Int).SetBytes(rBytes)
	s := new(big.Int).SetBytes(sBytes)

	return ecdsa.Verify(publicKey, payload, r, s), nil
}
