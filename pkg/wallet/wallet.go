package wallet

import (
	"chainnet/pkg/sign"
)

type Wallet struct {
	PrivateKey []byte
	PublicKey  []byte
}

func NewWallet(signature sign.Signature) *Wallet {
	privateKey, publicKey, err := signature.NewKeyPair()
	if err != nil {

	}

	return &Wallet{
		PrivateKey: privateKey,
		PublicKey:  publicKey,
	}
}

func (w *Wallet) GetAddress() []byte {
	return []byte{}
}
