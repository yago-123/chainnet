package errs

import "github.com/pkg/errors"

// Errors used in the storage package
var (
	ErrStorageElementNotFound = errors.New("not found")
)

// Errors used in the wallet package
var (
	ErrWalletInvalidChildPrivateKey = errors.New("derived private key is invalid")
)

// Errors used in the crypto package
var (
	ErrCryptoPublicKeyDerivation = errors.New("failed to derive public key from private key")
)

// Errors used in the mempool package
var (
	ErrMemPoolFull = errors.New("mempool does not have enough space")
)
