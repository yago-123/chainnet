package util

import (
	"chainnet/pkg/crypto/hash"
	"chainnet/pkg/kernel"
	"errors"
	"fmt"
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

func CalculateBlockHash(bh *kernel.BlockHeader, hasher hash.Hashing) ([]byte, error) {
	return hasher.Hash(bh.Assemble())
}

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
