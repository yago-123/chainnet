package utilcrypto

import (
	"crypto/ecdsa"
	"crypto/hmac"
	"crypto/sha512"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"os"
)

func ConvertECDSAKeysToBytes(pubKey *ecdsa.PublicKey, privKey *ecdsa.PrivateKey) ([]byte, []byte, error) {
	publicKey, err := ConvertECDSAPubToBytes(pubKey)
	if err != nil {
		return []byte{}, []byte{}, err
	}

	privateKey, err := ConvertECDSAPrivToBytes(privKey)
	if err != nil {
		return []byte{}, []byte{}, err
	}

	return publicKey, privateKey, nil
}

func ConvertECDSAPrivToBytes(privKey *ecdsa.PrivateKey) ([]byte, error) {
	// convert the private key to ASN.1/DER encoded form
	return x509.MarshalECPrivateKey(privKey)
}

func ConvertECDSAPubToBytes(pubKey *ecdsa.PublicKey) ([]byte, error) {
	// convert the public key to ASN.1/DER encoded form
	return x509.MarshalPKIXPublicKey(pubKey)
}

func DeriveECDSAPubFromPrivate(privKey []byte) ([]byte, error) {
	privateKeyECDSA, err := ConvertBytesToECDSAPriv(privKey)
	if err != nil {
		return nil, fmt.Errorf("error converting private key: %w", err)
	}

	if privateKeyECDSA == nil {
		return nil, fmt.Errorf("private key is nil")
	}

	pubkey, err := ConvertECDSAPubToBytes(&privateKeyECDSA.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("error deriving public key: %w", err)
	}

	return pubkey, nil
}

func ConvertBytesToECDSAPriv(privKey []byte) (*ecdsa.PrivateKey, error) {
	// parse the DER encoded private key to get ecdsa.PrivateKey
	return x509.ParseECPrivateKey(privKey)
}

func ConvertBytesToECDSAPub(pubKey []byte) (*ecdsa.PublicKey, error) {
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

// ReadECDSAPemPrivateKey reads an ECDSA private key from a PEM file
func ReadECDSAPemPrivateKey(path string) ([]byte, error) {
	privateKeyBytes, err := ReadFile(path)
	if err != nil {
		return []byte{}, fmt.Errorf("error reading private key file: %w", err)
	}

	// decode the PEM block
	block, _ := pem.Decode(privateKeyBytes)
	if block == nil {
		return []byte{}, fmt.Errorf("failed to decode PEM block containing private key")
	}

	return block.Bytes, nil
}

// ReadECDSAPemPublicKeyBytes reads an ECDSA public key from a PEM file and returns the raw DER encoded bytes.
func ReadECDSAPemPublicKeyBytes(path string) ([]byte, error) {
	publicKeyBytes, err := ReadFile(path)
	if err != nil {
		return []byte{}, fmt.Errorf("error reading private key file: %w", err)
	}

	// decode the PEM block
	block, _ := pem.Decode(publicKeyBytes)
	if block == nil {
		return nil, errors.New("failed to decode PEM block containing public key")
	}

	// return the raw DER encoded public key bytes
	return block.Bytes, nil
}

func CalculateHMACSha512(key []byte, data []byte) ([]byte, error) {
	h := hmac.New(sha512.New, key)
	_, err := h.Write(data)
	if err != nil {
		return []byte{}, fmt.Errorf("error writing data to HMAC: %w", err)
	}

	return h.Sum(nil), nil
}

func ReadFile(path string) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return []byte{}, fmt.Errorf("error opening file: %w", err)
	}
	defer file.Close()

	privateKeyBytes, err := io.ReadAll(file)
	if err != nil {
		return []byte{}, fmt.Errorf("error reading file: %w", err)
	}

	return privateKeyBytes, nil
}
