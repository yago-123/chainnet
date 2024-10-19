package crypto

import (
	"github.com/yago-123/chainnet/pkg/crypto/hash"
	"github.com/yago-123/chainnet/pkg/crypto/sign"
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
	h, err := hs.hasher.Hash(payload)
	if err != nil {
		return []byte{}, err
	}
	return hs.signature.Sign(h, privKey)
}

func (hs *HashedSignature) Verify(signature []byte, payload []byte, pubKey []byte) (bool, error) {
	h, err := hs.hasher.Hash(payload)
	if err != nil {
		return false, err
	}

	return hs.signature.Verify(signature, h, pubKey)
}
