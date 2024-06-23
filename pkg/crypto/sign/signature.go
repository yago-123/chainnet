package sign

type Signature interface {
	NewKeyPair() ([]byte, []byte, error)
	Sign(payload []byte) []byte
	Verify(signature []byte, payload []byte) bool
}
