package util

import (
	"chainnet/pkg/crypto/hash"
	"chainnet/pkg/kernel"
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"os"
)

const (
	NumBitsInByte   = 8
	BiggestByteMask = 0xFF
)

// CalculateTxHash calculates the hash of a transaction
func CalculateTxHash(tx *kernel.Transaction, hasher hash.Hashing) ([]byte, error) {
	// todo(): move this to the NewTransaction function instead?
	return hasher.Hash(tx.Assemble())
}

// VerifyTxHash verifies the hash of a transaction
func VerifyTxHash(tx *kernel.Transaction, hash []byte, hasher hash.Hashing) error {
	ret, err := hasher.Verify(hash, tx.Assemble())
	if err != nil {
		return fmt.Errorf("verify tx hash failed: %w", err)
	}

	if !ret {
		return errors.New("tx hash verification failed")
	}

	return nil
}

// CalculateBlockHash calculates the hash of a block header
func CalculateBlockHash(bh *kernel.BlockHeader, hasher hash.Hashing) ([]byte, error) {
	return hasher.Hash(bh.Assemble())
}

// VerifyBlockHash verifies the hash of a block header
func VerifyBlockHash(bh *kernel.BlockHeader, hash []byte, hasher hash.Hashing) error {
	ret, err := hasher.Verify(hash, bh.Assemble())
	if err != nil {
		return fmt.Errorf("block hashing failed: %w", err)
	}

	if !ret {
		return errors.New("block header hash verification failed")
	}

	return nil
}

// IsFirstNBitsZero checks if the first n bits of the array are zero
func IsFirstNBitsZero(arr []byte, n uint) bool {
	// calculate the number of full bytes and remaining bits
	fullBytes := n / NumBitsInByte
	remainingBits := n % NumBitsInByte

	// check if the array is long enough to contain n bits
	if len(arr) < int(fullBytes) {
		return false
	}

	// check full bytes
	for _, b := range arr[:int(fullBytes)] {
		if b != 0 {
			return false
		}
	}

	// check remaining bits in the next byte if there are any
	if remainingBits > 0 {
		nextByte := arr[int(fullBytes)]
		// create a mask to isolate the remaining bits
		mask := byte(BiggestByteMask << (NumBitsInByte - remainingBits))
		if nextByte&mask != 0 {
			return false
		}
	}

	return true
}

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
	privateKeyBytes, err := readFile(path)
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
	publicKeyBytes, err := readFile(path)
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

func readFile(path string) ([]byte, error) {
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
