package hd

import (
	"bytes"
	"encoding/gob"
	"fmt"
)

type HDAccountMetadata struct {
	WalletIndex uint32
}

type HDMetadata struct {
	AccountIndex uint32
	Accounts     []HDAccountMetadata
}

func SerializeHDMetadata(metadata *HDMetadata) ([]byte, error) {
	var buf bytes.Buffer

	encoder := gob.NewEncoder(&buf)
	if err := encoder.Encode(metadata); err != nil {
		return []byte{}, fmt.Errorf("error encoding wallet metadata: %w", err)
	}

	return buf.Bytes(), nil
}

func DeserializeHDMetadata(data []byte) (*HDMetadata, error) {
	var decodedMetadata *HDMetadata

	decoder := gob.NewDecoder(bytes.NewReader(data))
	if err := decoder.Decode(&decodedMetadata); err != nil {
		return nil, fmt.Errorf("error decoding wallet metadata: %w", err)
	}

	return decodedMetadata, nil
}
