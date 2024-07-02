package sign

type Signature interface {
	NewKeyPair() ([]byte, []byte, error)
	// todo() should we add the transaction object here directly instead of payload?
	Sign(payload []byte, privKey []byte) ([]byte, error)
	Verify(signature []byte, payload []byte, pubKey []byte) (bool, error)
}
