package sign

type Signature interface {
	NewKeyPair() ([]byte, []byte, error)
	Sign([]byte) ([]byte, error)
	Verify([]byte, []byte) error
}
