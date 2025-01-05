package hd

type HDAccountMetadata struct {
}

type HDMetadata struct {
	HDAccountMetadata
}

func NewHDMetadata() *HDMetadata {
	return &HDMetadata{}
}
