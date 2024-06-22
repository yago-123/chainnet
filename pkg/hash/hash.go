package hash

type Hashing interface {
	Hash(payload []byte) []byte
	Verify(hash []byte, payload []byte) bool
}
