package util_p2pkh

import (
	"bytes"
	"fmt"
	"github.com/yago-123/chainnet/pkg/crypto"
	"github.com/yago-123/chainnet/pkg/crypto/hash"
)

const (
	P2PKHAddressLength    = 1 + 20 + 4 // version + pubKeyHash + checksum
	P2PKHPubKeyHashLength = 20
)

// GenerateP2PKHAddrFromPubKey generates a P2PKH address from a public key (including a checksum for error detection).
// Returns the P2PKH address as a base58 encoded string.
func GenerateP2PKHAddrFromPubKey(pubKey []byte, version byte) ([]byte, error) {
	hasherP2PKH := crypto.NewMultiHash([]hash.Hashing{hash.NewSHA256(), hash.NewRipemd160()})

	// hash the public key
	pubKeyHash, err := hasherP2PKH.Hash(pubKey)
	if err != nil {
		return []byte{}, fmt.Errorf("could not hash the public key: %w", err)
	}

	// add the version to the hashed public key in order to hash again and obtain the checksum
	versionedPayload := append([]byte{version}, pubKeyHash...)
	// todo() checksum must be a double SHA-256 hash, instead of SHA-256 + RIPEMD-160, but for now is OK
	checksum, err := hasherP2PKH.Hash(versionedPayload)
	if err != nil {
		return []byte{}, fmt.Errorf("could not hash the versioned payload: %w", err)
	}

	// add checksum to generate address
	payload := append(versionedPayload, checksum[:4]...) //nolint:gocritic // we need to append the checksum to the payload

	return payload, nil
}

// ExtractPubKeyHashedFromP2PKHAddr extracts the public key hash from a P2PKH address
func ExtractPubKeyHashedFromP2PKHAddr(address []byte) ([]byte, byte, error) {
	hasherP2PKH := crypto.NewMultiHash([]hash.Hashing{hash.NewSHA256(), hash.NewRipemd160()})

	// check that address has at least the minimum valid length (1 version + 1 pubKeyHash + 4 checksum)
	// we know that the address should be 20 bytes because it is a RIPEMD hash, but for now this is OK
	if len(address) != P2PKHAddressLength {
		return nil, 0, fmt.Errorf("invalid P2PKH address length: got %d, want at least %d", len(address), P2PKHAddressLength)
	}

	version := address[0]

	// extract the public key hash (remaining bytes except for the last 4, if available)
	pubKeyHash := address[1 : len(address)-4]

	// Ensure that the public key hash is not empty
	if len(pubKeyHash) != P2PKHPubKeyHashLength {
		return nil, 0, fmt.Errorf("invalid public key hash length: got %d, want %d", len(pubKeyHash), P2PKHPubKeyHashLength)
	}

	// verify the checksum
	checksum := address[len(address)-4:]
	err := verifyP2PKHChecksum(version, pubKeyHash, checksum, hasherP2PKH)
	if err != nil {
		return nil, 0, err
	}

	return pubKeyHash, version, nil
}

func verifyP2PKHChecksum(version byte, pubKeyHash, checksum []byte, hasherP2PKH hash.Hashing) error {
	versionPayload := append([]byte{version}, pubKeyHash...)
	calculatedChecksum, err := hasherP2PKH.Hash(versionPayload)
	if err != nil {
		return fmt.Errorf("could not hash the versioned payload: %w", err)
	}

	if !bytes.Equal(checksum, calculatedChecksum[:4]) {
		return fmt.Errorf("error validating checksum, expected %x, got %x", checksum, calculatedChecksum[:4])
	}

	return nil
}
