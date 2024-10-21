package encoding

import (
	"bytes"
	"encoding/gob"
	"fmt"

	"github.com/yago-123/chainnet/pkg/kernel"
)

type GobEncoder struct {
}

func NewGobEncoder() *GobEncoder {
	return &GobEncoder{}
}

func (gobenc *GobEncoder) SerializeBlock(b kernel.Block) ([]byte, error) {
	var result bytes.Buffer

	encoder := gob.NewEncoder(&result)
	if err := encoder.Encode(b); err != nil {
		return []byte{}, fmt.Errorf("error serializing block %x: %w", b.Hash, err)
	}

	return result.Bytes(), nil
}

func (gobenc *GobEncoder) DeserializeBlock(data []byte) (*kernel.Block, error) {
	var b kernel.Block

	decoder := gob.NewDecoder(bytes.NewReader(data))
	if err := decoder.Decode(&b); err != nil {
		return &kernel.Block{}, fmt.Errorf("error deserializing block: %w", err)
	}

	return &b, nil
}

func (gobenc *GobEncoder) SerializeHeader(bh kernel.BlockHeader) ([]byte, error) {
	var result bytes.Buffer

	encoder := gob.NewEncoder(&result)
	if err := encoder.Encode(bh); err != nil {
		return []byte{}, fmt.Errorf("error serializing block header: %w", err)
	}

	return result.Bytes(), nil
}

func (gobenc *GobEncoder) DeserializeHeader(data []byte) (*kernel.BlockHeader, error) {
	var bh kernel.BlockHeader

	decoder := gob.NewDecoder(bytes.NewReader(data))
	if err := decoder.Decode(&bh); err != nil {
		return &kernel.BlockHeader{}, fmt.Errorf("error deserializing block header: %w", err)
	}

	return &bh, nil
}

func (gobenc *GobEncoder) SerializeHeaders(headers []*kernel.BlockHeader) ([]byte, error) {
	var result bytes.Buffer

	encoder := gob.NewEncoder(&result)
	if err := encoder.Encode(headers); err != nil {
		return nil, fmt.Errorf("error serializing block headers: %w", err)
	}

	return result.Bytes(), nil
}

func (gobenc *GobEncoder) DeserializeHeaders(data []byte) ([]*kernel.BlockHeader, error) {
	var headers []*kernel.BlockHeader

	decoder := gob.NewDecoder(bytes.NewReader(data))
	if err := decoder.Decode(&headers); err != nil {
		return nil, fmt.Errorf("error deserializing block headers: %w", err)
	}

	return headers, nil
}

func (gobenc *GobEncoder) SerializeTransaction(tx kernel.Transaction) ([]byte, error) {
	var result bytes.Buffer

	enc := gob.NewEncoder(&result)
	if err := enc.Encode(tx); err != nil {
		return []byte{}, fmt.Errorf("error serializing transaction %s: %w", string(tx.ID), err)
	}

	return result.Bytes(), nil
}

func (gobenc *GobEncoder) DeserializeTransaction(data []byte) (*kernel.Transaction, error) {
	var tx *kernel.Transaction

	decoder := gob.NewDecoder(bytes.NewReader(data))
	if err := decoder.Decode(&tx); err != nil {
		return &kernel.Transaction{}, fmt.Errorf("error deserializing transaction: %w", err)
	}

	return tx, nil
}

func (gobenc *GobEncoder) SerializeTransactions(txs []*kernel.Transaction) ([]byte, error) {
	var result bytes.Buffer

	encoder := gob.NewEncoder(&result)
	if err := encoder.Encode(txs); err != nil {
		return nil, fmt.Errorf("error serializing transactions: %w", err)
	}

	return result.Bytes(), nil
}

func (gobenc *GobEncoder) DeserializeTransactions(data []byte) ([]*kernel.Transaction, error) {
	var txs []*kernel.Transaction

	decoder := gob.NewDecoder(bytes.NewReader(data))
	if err := decoder.Decode(&txs); err != nil {
		return []*kernel.Transaction{}, fmt.Errorf("error deserializing transactions: %w", err)
	}

	return txs, nil
}

func (gobenc *GobEncoder) SerializeUTXO(utxo kernel.UTXO) ([]byte, error) {
	var result bytes.Buffer

	encoder := gob.NewEncoder(&result)
	if err := encoder.Encode(utxo); err != nil {
		return nil, fmt.Errorf("error serializing UTXO %x: %w", utxo.TxID, err)
	}

	return result.Bytes(), nil
}

func (gobenc *GobEncoder) DeserializeUTXO(data []byte) (*kernel.UTXO, error) {
	var utxo *kernel.UTXO

	decoder := gob.NewDecoder(bytes.NewReader(data))
	if err := decoder.Decode(utxo); err != nil {
		return &kernel.UTXO{}, fmt.Errorf("error deserializing UTXO: %w", err)
	}

	return utxo, nil
}

func (gobenc *GobEncoder) SerializeUTXOs(utxos []*kernel.UTXO) ([]byte, error) {
	var result bytes.Buffer

	encoder := gob.NewEncoder(&result)
	if err := encoder.Encode(utxos); err != nil {
		return nil, fmt.Errorf("error serializing UTXOs: %w", err)
	}

	return result.Bytes(), nil
}

func (gobenc *GobEncoder) DeserializeUTXOs(data []byte) ([]*kernel.UTXO, error) {
	var utxos []*kernel.UTXO

	decoder := gob.NewDecoder(bytes.NewReader(data))
	if err := decoder.Decode(&utxos); err != nil {
		return nil, fmt.Errorf("error deserializing UTXOs: %w", err)
	}

	return utxos, nil
}
