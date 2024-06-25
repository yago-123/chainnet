package crypto

import (
	"chainnet/pkg/crypto/hash"
	"chainnet/pkg/crypto/sign"
)

type HashedSignature struct {
	signature sign.Signature
	hasher    hash.Hashing
}

func NewHashedSignature(signature sign.Signature, hasher hash.Hashing) *HashedSignature {
	return &HashedSignature{
		signature: signature,
		hasher:    hasher,
	}
}

func (hs *HashedSignature) NewKeyPair() ([]byte, []byte, error) {
	return hs.signature.NewKeyPair()
}

func (hs *HashedSignature) Sign(payload []byte, privKey []byte) ([]byte, error) {
	return hs.signature.Sign(hs.hasher.Hash(payload), privKey)
}

func (hs *HashedSignature) Verify(signature []byte, payload []byte, pubKey []byte) (bool, error) {
	return hs.signature.Verify(hs.hasher.Hash(payload), signature, pubKey)
}
