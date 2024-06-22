package wallet

import (
	"chainnet/pkg/crypto/hash"
	"chainnet/pkg/crypto/sign"
	"github.com/btcsuite/btcutil/base58"
)

type Wallet struct {
	privateKey []byte
	PublicKey  []byte
	hasher     hash.Hashing
	version    []byte
}

// NewWallet (ecdsa, NewMultiHash(NewSHA256, NewRipemd())))
func NewWallet(version []byte, signature sign.Signature, hashing hash.Hashing) (*Wallet, error) {
	privateKey, publicKey, err := signature.NewKeyPair()
	if err != nil {
		return nil, err
	}

	return &Wallet{
		privateKey: privateKey,
		PublicKey:  publicKey,
		hasher:     hashing,
		version:    version,
	}, nil
}

// GetAddress returns one wallet address
// todo() implement hierarchically deterministic HD wallet
func (w *Wallet) GetAddress() []byte {
	// hash the public key
	pubKeyHash := w.hasher.Hash(w.PublicKey)

	// add the version to the hashed public key in order to hash again and obtain the checksum
	versionedPayload := append(w.version, pubKeyHash...)
	checksum := w.hasher.Hash(versionedPayload)

	// return the base58 of the versioned payload and the checksum
	payload := append(versionedPayload, checksum...)
	return []byte(base58.Encode(payload))
}
