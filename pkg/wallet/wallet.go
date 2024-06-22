package wallet

import (
	"chainnet/pkg/crypto/hash"
	"chainnet/pkg/crypto/sign"
	base58 "github.com/btcsuite/btcutil/base58"
)

type Wallet struct {
	PrivateKey []byte
	PublicKey  []byte
	hasher     hash.Hashing
	version    []byte
}

func NewWallet(version []byte, signature sign.Signature) *Wallet {
	privateKey, publicKey, err := signature.NewKeyPair()
	if err != nil {

	}

	// todo() this should be passed by argument in NewWallet
	multiHash, err := hash.NewMultiHash([]hash.Hashing{hash.NewSHA256(), hash.NewRipemd160()})
	if err != nil {

	}

	return &Wallet{
		PrivateKey: privateKey,
		PublicKey:  publicKey,
		hasher:     multiHash,
		version:    version,
	}
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
