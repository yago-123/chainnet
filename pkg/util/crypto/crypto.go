package utilcrypto

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/hmac"
	"crypto/sha512"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"math/big"
	"os"
)

const (
	// length of P-256 curve for private key
	Secp256r1KeyLength = 32
)

// ConvertECDSAKeysToDERBytes converts ECDSA public and private keys to DER encoded byte arrays
func ConvertECDSAKeysToDERBytes(pubKey *ecdsa.PublicKey, privKey *ecdsa.PrivateKey) ([]byte, []byte, error) {
	publicKey, err := ConvertECDSAPubToDERBytes(pubKey)
	if err != nil {
		return []byte{}, []byte{}, err
	}

	privateKey, err := ConvertECDSAPrivToDERBytes(privKey)
	if err != nil {
		return []byte{}, []byte{}, err
	}

	return publicKey, privateKey, nil
}

// ConvertECDSAPrivToDERBytes converts an ECDSA private key to a DER encoded byte array
func ConvertECDSAPrivToDERBytes(privKey *ecdsa.PrivateKey) ([]byte, error) {
	// convert the private key to ASN.1/DER encoded form
	return x509.MarshalECPrivateKey(privKey)
}

// ConvertECDSAPubToDERBytes converts an ECDSA public key to a DER encoded byte array
func ConvertECDSAPubToDERBytes(pubKey *ecdsa.PublicKey) ([]byte, error) {
	// convert the public key to ASN.1/DER encoded form
	return x509.MarshalPKIXPublicKey(pubKey)
}

// DeriveECDSAPubFromPrivateDERBytes derives a DER public key array from a DER encoded private key
func DeriveECDSAPubFromPrivateDERBytes(privKey []byte) ([]byte, error) {
	privateKeyECDSA, err := ConvertDERBytesToECDSAPriv(privKey)
	if err != nil {
		return nil, fmt.Errorf("error converting private key: %w", err)
	}

	if privateKeyECDSA == nil {
		return nil, fmt.Errorf("private key is nil")
	}

	pubkey, err := ConvertECDSAPubToDERBytes(&privateKeyECDSA.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("error deriving public key: %w", err)
	}

	return pubkey, nil
}

// ConvertDERBytesToECDSAPriv converts a DER encoded private key to an ECDSA private key
func ConvertDERBytesToECDSAPriv(privKey []byte) (*ecdsa.PrivateKey, error) {
	// parse the DER encoded private key to get ecdsa.PrivateKey
	return x509.ParseECPrivateKey(privKey)
}

// ConvertDERBytesToECDSAPub converts a DER encoded public key to an ECDSA public key
func ConvertDERBytesToECDSAPub(pubKey []byte) (*ecdsa.PublicKey, error) {
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

// ReadECDSAPemToPrivateKeyDerBytes reads an ECDSA private key from a PEM file
func ReadECDSAPemToPrivateKeyDerBytes(path string) ([]byte, error) {
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

// EncodeRawPrivateKeyToDERBytes encodes a private key to DER format. DER is the binary format used for managing keys
// because the portability of the keys across different systems is valued in this project
func EncodeRawPrivateKeyToDERBytes(privKey []byte) ([]byte, error) {
	// todo(): include support more crypto curves for ECDSA
	if len(privKey) != Secp256r1KeyLength {
		return nil, fmt.Errorf("invalid private key length: only P-256 curve (32 bytes) is supported so far, got %d", len(privKey))
	}

	// create the ECDSA private key structure
	privKeyECDSA := new(ecdsa.PrivateKey)
	privKeyECDSA.PublicKey.Curve = elliptic.P256()  // by default we use P256 for generating private keys
	privKeyECDSA.D = new(big.Int).SetBytes(privKey) // set the private key as a big integer

	// generate the public key from the private key
	privKeyECDSA.PublicKey.X, privKeyECDSA.PublicKey.Y = privKeyECDSA.PublicKey.Curve.ScalarBaseMult(privKeyECDSA.D.Bytes())

	// encode the private key to DER format
	privKeyDER, err := x509.MarshalECPrivateKey(privKeyECDSA)
	if err != nil {
		return nil, fmt.Errorf("error marshaling private key to DER: %w", err)
	}

	return privKeyDER, nil
}

// DecodeDERBytesToRawPrivateKey decodes a DER encoded private key to raw bytes. DER is the binary format used for
// managing keys because the portability of the keys across different systems is valued in this project
func DecodeDERBytesToRawPrivateKey(derPrivateBytes []byte) ([]byte, error) {
	// parse the DER-encoded private key
	privKeyECDSA, err := x509.ParseECPrivateKey(derPrivateBytes)
	if err != nil {
		return nil, fmt.Errorf("error parsing DER bytes: %w", err)
	}

	// ensure the curve is P-256, so far is the only curve supported
	// todo(): implement support for more curves in ECDSA
	if privKeyECDSA.Curve != elliptic.P256() {
		return nil, fmt.Errorf("unsupported curve: only P-256 is supported")
	}

	// return the raw private key bytes
	return privKeyECDSA.D.Bytes(), nil
}

// DecodeDERBytesToRawPublicKey decodes a DER-encoded public key to raw bytes. Only P-256 curve is supported for now
func DecodeDERBytesToRawPublicKey(derPublicBytes []byte) ([]byte, error) {
	// parse the DER-encoded public key
	pubKey, err := x509.ParsePKIXPublicKey(derPublicBytes)
	if err != nil {
		return nil, fmt.Errorf("error parsing DER bytes: %w", err)
	}

	// assert the parsed public key is an ECDSA public key
	pubKeyECDSA, ok := pubKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("key is not an ECDSA public key")
	}

	// ensure the curve is P-256, so far is the only curve supported
	// todo(): implement support for more curves in ECDSA
	if pubKeyECDSA.Curve != elliptic.P256() {
		return nil, fmt.Errorf("unsupported curve: only P-256 is supported")
	}

	// concatenate X and Y coordinates to form the raw public key bytes
	rawPubKey := append(pubKeyECDSA.X.Bytes(), pubKeyECDSA.Y.Bytes()...)

	return rawPubKey, nil
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

	fileContent, err := io.ReadAll(file)
	if err != nil {
		return []byte{}, fmt.Errorf("error reading file: %w", err)
	}

	return fileContent, nil
}
