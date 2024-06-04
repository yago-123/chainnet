package persistence

type Persistence interface {
	SerializeBlock(b Block) []byte
}
