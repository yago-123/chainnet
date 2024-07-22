package hash

import (
	"fmt"
)

type Void struct {
}

func NewVoidHasher() *Void {
	return &Void{}
}

func (v *Void) Hash(payload []byte) ([]byte, error) {
	return []byte{}, fmt.Errorf("void hasher does not hash anything")
}

func (v *Void) Verify(hash []byte, payload []byte) (bool, error) {
	return false, fmt.Errorf("void hasher does not verify anything")
}
