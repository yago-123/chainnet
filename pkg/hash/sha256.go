package hash

import (
	"crypto/sha256"
	"reflect"
)

type Sha256 struct {
}

func NewSHA256() *Sha256 {
	return &Sha256{}
}

func (h *Sha256) Hash(payload []byte) []byte {
	ret := sha256.Sum256(payload)
	return ret[:]
}

func (h *Sha256) Verify(hash []byte, payload []byte) bool {
	return reflect.DeepEqual(hash, h.Hash(payload))
}
