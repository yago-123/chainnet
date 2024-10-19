package encoding_test

import (
	"bytes"
	"encoding/gob"
	"testing"

	"github.com/yago-123/chainnet/pkg/encoding"
	"github.com/yago-123/chainnet/pkg/kernel"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testBlock = &kernel.Block{ //nolint:gochecknoglobals // data that is used across all test funcs
	Header: &kernel.BlockHeader{
		Version:       []byte("v1"),
		PrevBlockHash: []byte("prevhash"),
		MerkleRoot:    []byte("merkleroot"),
		Height:        123,
		Timestamp:     1610000000,
		Target:        456,
		Nonce:         789,
	},
	Transactions: []*kernel.Transaction{
		{
			ID: []byte("tx1"),
			Vin: []kernel.TxInput{
				{Txid: []byte("tx0"), Vout: 0, ScriptSig: "script1", PubKey: "pubkey1"},
				{Txid: []byte("tx0"), Vout: 1, ScriptSig: "script2", PubKey: "pubkey2"},
			},
			Vout: []kernel.TxOutput{
				{Amount: 50, ScriptPubKey: "scriptpubkey1", PubKey: "pubkey1"},
				{Amount: 30, ScriptPubKey: "scriptpubkey2", PubKey: "pubkey2"},
			},
		},
		{
			ID: []byte("tx2"),
			Vin: []kernel.TxInput{
				{Txid: []byte("tx1"), Vout: 0, ScriptSig: "script3", PubKey: "pubkey3"},
			},
			Vout: []kernel.TxOutput{
				{Amount: 20, ScriptPubKey: "scriptpubkey3", PubKey: "pubkey3"},
				{Amount: 10, ScriptPubKey: "scriptpubkey4", PubKey: "pubkey4"},
			},
		},
	},
	Hash: []byte("blockhash"),
}

var testTransaction = kernel.Transaction{ //nolint:gochecknoglobals // data that is used across all test funcs
	ID: []byte("tx1"),
	Vin: []kernel.TxInput{
		{Txid: []byte("tx0"), Vout: 0, ScriptSig: "script1", PubKey: "pubkey1"},
		{Txid: []byte("tx0"), Vout: 1, ScriptSig: "script2", PubKey: "pubkey2"},
	},
	Vout: []kernel.TxOutput{
		{Amount: 50, ScriptPubKey: "scriptpubkey1", PubKey: "pubkey1"},
		{Amount: 30, ScriptPubKey: "scriptpubkey2", PubKey: "pubkey2"},
	},
}

var testBlockHeaders = []*kernel.BlockHeader{ //nolint:gochecknoglobals // data that is used across all test funcs
	{
		Version:       []byte("v1"),
		PrevBlockHash: []byte("prevhash"),
		MerkleRoot:    []byte("merkleroot"),
		Height:        123,
		Timestamp:     1610000000,
		Target:        456,
		Nonce:         789,
	},
	{
		Version:       []byte("v2"),
		PrevBlockHash: []byte("prevhash2"),
		MerkleRoot:    []byte("merkleroot2"),
		Height:        456,
		Timestamp:     1620000000,
		Target:        789,
		Nonce:         101112,
	},
}

func TestSerializeBlock(t *testing.T) {
	gobenc := encoding.NewGobEncoder()

	data, err := gobenc.SerializeBlock(*testBlock)
	require.NoError(t, err)

	var block kernel.Block
	decoder := gob.NewDecoder(bytes.NewReader(data))
	err = decoder.Decode(&block)
	require.NoError(t, err)

	assert.Equal(t, *testBlock, block)
}

func TestDeserializeBlock(t *testing.T) {
	gobenc := encoding.NewGobEncoder()

	data, err := gobenc.SerializeBlock(*testBlock)
	require.NoError(t, err)

	block, err := gobenc.DeserializeBlock(data)
	require.NoError(t, err)

	assert.Equal(t, testBlock, block)
}

func TestSerializeHeader(t *testing.T) {
	gobenc := encoding.NewGobEncoder()

	data, err := gobenc.SerializeHeader(*testBlock.Header)
	require.NoError(t, err)

	var header kernel.BlockHeader
	decoder := gob.NewDecoder(bytes.NewReader(data))
	err = decoder.Decode(&header)
	require.NoError(t, err)

	assert.Equal(t, *testBlock.Header, header)
}

func TestDeserializeHeader(t *testing.T) {
	gobenc := encoding.NewGobEncoder()

	data, err := gobenc.SerializeHeader(*testBlock.Header)
	require.NoError(t, err)

	header, err := gobenc.DeserializeHeader(data)
	require.NoError(t, err)

	assert.Equal(t, testBlock.Header, header)
}

func TestSerializeHeaders(t *testing.T) {
	gobenc := encoding.NewGobEncoder()

	data, err := gobenc.SerializeHeaders(testBlockHeaders)
	require.NoError(t, err)

	var headers []*kernel.BlockHeader
	decoder := gob.NewDecoder(bytes.NewReader(data))
	err = decoder.Decode(&headers)
	require.NoError(t, err)

	assert.ElementsMatch(t, testBlockHeaders, headers)
}

func TestDeserializeHeaders(t *testing.T) {
	gobenc := encoding.NewGobEncoder()

	data, err := gobenc.SerializeHeaders(testBlockHeaders)
	require.NoError(t, err)

	headers, err := gobenc.DeserializeHeaders(data)
	require.NoError(t, err)

	assert.ElementsMatch(t, testBlockHeaders, headers)
}

func TestSerializeTransaction(t *testing.T) {
	gobenc := encoding.NewGobEncoder()

	data, err := gobenc.SerializeTransaction(testTransaction)
	require.NoError(t, err)

	var transaction kernel.Transaction
	decoder := gob.NewDecoder(bytes.NewReader(data))
	err = decoder.Decode(&transaction)
	require.NoError(t, err)

	assert.Equal(t, testTransaction, transaction)
}

func TestDeserializeTransaction(t *testing.T) {
	gobenc := encoding.NewGobEncoder()

	data, err := gobenc.SerializeTransaction(testTransaction)
	require.NoError(t, err)

	transaction, err := gobenc.DeserializeTransaction(data)
	require.NoError(t, err)

	assert.Equal(t, &testTransaction, transaction)
}
