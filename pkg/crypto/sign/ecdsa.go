package sign

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"errors"
	"math/big"
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

	return convertToBytes(&private.PublicKey, private)
}

func (ecdsaSign *ECDSASigner) Sign(payload []byte, privKey []byte) ([]byte, error) {
	if err := signInputValidator(payload, privKey); err != nil {
		return []byte{}, err
	}

	privateKey, err := convertBytesToPrivateKey(privKey)
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

	publicKey, err := convertBytesToPublicKey(pubKey)
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

func convertToBytes(pubKey *ecdsa.PublicKey, privKey *ecdsa.PrivateKey) ([]byte, []byte, error) {
	// convert the public key to ASN.1/DER encoded form
	publicKey, err := x509.MarshalPKIXPublicKey(pubKey)
	if err != nil {
		return []byte{}, []byte{}, err
	}

	// convert the private key to ASN.1/DER encoded form
	privateKey, err := x509.MarshalECPrivateKey(privKey)
	if err != nil {
		return []byte{}, []byte{}, err
	}

	return publicKey, privateKey, nil
}

func convertBytesToPrivateKey(privKey []byte) (*ecdsa.PrivateKey, error) {
	// parse the DER encoded private key to get ecdsa.PrivateKey
	return x509.ParseECPrivateKey(privKey)
}

func convertBytesToPublicKey(pubKey []byte) (*ecdsa.PublicKey, error) {
	// parse the DER encoded public key to get ecdsa.PublicKey
	pub, err := x509.ParsePKIXPublicKey(pubKey)
	if err != nil {
		return nil, err
	}

	publicKey, ok := pub.(*ecdsa.PublicKey)
	if !ok {
		return nil, errors.New("error deserializing ECDSA public key")
	}

	return publicKey, nil
}
