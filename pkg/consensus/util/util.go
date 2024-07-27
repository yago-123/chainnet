package util

import (
	"chainnet/pkg/crypto/hash"
	"chainnet/pkg/kernel"
	"errors"
	"fmt"
)

// CalculateTxHash calculates the hash of a transaction
func CalculateTxHash(tx *kernel.Transaction, hasher hash.Hashing) ([]byte, error) {
	// todo(): move this to the NewTransaction function instead?
	return hasher.Hash(tx.Assemble())
}

func VerifyTxHash(tx *kernel.Transaction, hash []byte, hasher hash.Hashing) error {
	ret, err := hasher.Verify(hash, tx.Assemble())
	if err != nil {
		return fmt.Errorf("verify tx hash failed: %v", err)
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
		return fmt.Errorf("block hashing failed: %v", err)
	}

	if !ret {
		return errors.New("block header hash verification failed")
	}

	return nil
}
