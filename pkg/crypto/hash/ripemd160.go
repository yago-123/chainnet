package hash

import (
	"bytes"
	"golang.org/x/crypto/ripemd160" //nolint
	"hash"
)

type Ripemd160 struct {
	ripe hash.Hash
}

func NewRipemd160() *Ripemd160 {
	return &Ripemd160{ripe: ripemd160.New()}
}

func (r *Ripemd160) Hash(payload []byte) []byte {
	return r.ripe.Sum(payload)
}

func (r *Ripemd160) Verify(hash []byte, payload []byte) bool {
	return bytes.Equal(hash, r.Hash(payload))
}
