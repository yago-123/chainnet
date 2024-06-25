package wallet

import (
	"chainnet/pkg/crypto/hash"
	"chainnet/pkg/crypto/sign"
	"chainnet/pkg/kernel"
	base58 "github.com/btcsuite/btcutil/base58"
)

type Wallet struct {
	version    []byte
	privateKey []byte
	publicKey  []byte
	signer     sign.Signature
	hasher     hash.Hashing
}

func (w *Wallet) ID() string {
	return string(w.hasher.Hash(w.publicKey))
}

func NewWallet(version []byte, signer sign.Signature, hasher hash.Hashing) (*Wallet, error) {
	privateKey, publicKey, err := signer.NewKeyPair()
	if err != nil {
		return nil, err
	}

	return &Wallet{
		privateKey: privateKey,
		publicKey:  publicKey,
		signer:     signer,
		hasher:     hasher,
		version:    version,
	}, nil
}

// GetAddress returns one wallet address
// todo() implement hierarchically deterministic HD wallet
func (w *Wallet) GetAddress() []byte {
	// hash the public key
	pubKeyHash := w.hasher.Hash(w.publicKey)

	// add the version to the hashed public key in order to hash again and obtain the checksum
	versionedPayload := append(w.version, pubKeyHash...)
	checksum := w.hasher.Hash(versionedPayload)

	// return the base58 of the versioned payload and the checksum
	payload := append(versionedPayload, checksum...)
	return []byte(base58.Encode(payload))
}

// UnlockTxFunds take a tx that is being built and unlocks the UTXOs from which the input funds are going to
// be used
func (w *Wallet) UnlockTxFunds(tx *kernel.Transaction) (*kernel.Transaction, error) {

	// todo() for now, this only applies to P2PK, be able to extend once pkg/script/interpreter.go is created
	// todo() we must also have access to the previous tx output in order to verify the ScriptPubKey script
	txData := tx.AssembleForSigning()

	for _, vin := range tx.Vin {
		if vin.CanUnlockOutputWith(string(w.publicKey)) {
			signature, err := w.signer.Sign(txData, w.privateKey)
			if err != nil {
				return nil, err
			}

			vin.ScriptSig = string(signature)
		}
	}

	return tx, nil
}
