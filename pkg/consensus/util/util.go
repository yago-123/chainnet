package util

import (
	"chainnet/pkg/crypto/hash"
	"chainnet/pkg/kernel"
)

// CalculateTxHash calculates the hash of a transaction
func CalculateTxHash(tx *kernel.Transaction, hasher hash.Hashing) ([]byte, error) {
	// todo(): move this to the NewTransaction function instead?
	return hasher.Hash(tx.Assemble())
}
