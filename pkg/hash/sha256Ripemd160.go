package hash

type SHA256Ripemd160 struct {
}

func NewSHA256Ripemd160() *SHA256Ripemd160 {
	return &SHA256Ripemd160{}
}

func (hash *SHA256Ripemd160) Hash(payload []byte) []byte {
	return []byte{}
}

func (hash *SHA256Ripemd160) CheckSum(payload []byte) []byte {
	return []byte{}
}
