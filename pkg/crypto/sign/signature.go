package sign

type Signature interface {
	NewKeyPair() ([]byte, []byte, error)
	Sign(payload []byte, privKey []byte) ([]byte, error)
	Verify(signature []byte, payload []byte, pubKey []byte) (bool, error)
}
