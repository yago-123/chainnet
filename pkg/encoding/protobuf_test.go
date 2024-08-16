package encoding //nolint:testpackage // don't create separate package for tests

import (
	pb "chainnet/pkg/chain/p2p/protobuf"
	"chainnet/pkg/crypto/sign"
	"chainnet/pkg/kernel"
	"fmt"
	"github.com/stretchr/testify/require"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

var testBlock = kernel.Block{ //nolint:gochecknoglobals // ignore linter in this case
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
			ID: []byte("txid1"),
			Vin: []kernel.TxInput{
				{
					Txid:      []byte("txid0"),
					Vout:      0,
					ScriptSig: "sig1",
					PubKey:    "pubkey1",
				},
			},
			Vout: []kernel.TxOutput{
				{
					Amount:       100,
					ScriptPubKey: "scriptpubkey1",
					PubKey:       "pubkey1",
				},
			},
		},
	},
	Hash: []byte("blockhash"),
}

var expectedPbBlock = &pb.Block{ //nolint:gochecknoglobals // ignore linter in this case
	Header: &pb.BlockHeader{
		Version:       []byte("v1"),
		PrevBlockHash: []byte("prevhash"),
		MerkleRoot:    []byte("merkleroot"),
		Height:        123,
		Timestamp:     1610000000,
		Target:        456,
		Nonce:         789,
	},
	Transactions: []*pb.Transaction{
		{
			Id: []byte("txid1"),
			Vin: []*pb.TxInput{
				{
					Txid:      []byte("txid0"),
					Vout:      0,
					ScriptSig: "sig1",
					PubKey:    "pubkey1",
				},
			},
			Vout: []*pb.TxOutput{
				{
					Amount:       100,
					ScriptPubKey: "scriptpubkey1",
					PubKey:       "pubkey1",
				},
			},
		},
	},
	Hash: []byte("blockhash"),
}

var testBlockHeader = &kernel.BlockHeader{ //nolint:gochecknoglobals // data that is used across all test funcs
	Version:       []byte("v1"),
	PrevBlockHash: []byte("prevhash"),
	MerkleRoot:    []byte("merkleroot"),
	Height:        123,
	Timestamp:     1610000000,
	Target:        456,
	Nonce:         789,
}

var expectedPbBlockHeader = &pb.BlockHeader{ //nolint:gochecknoglobals // data that is used across all test funcs
	Version:       []byte("v1"),
	PrevBlockHash: []byte("prevhash"),
	MerkleRoot:    []byte("merkleroot"),
	Height:        123,
	Timestamp:     1610000000,
	Target:        456,
	Nonce:         789,
}

var testBlockHeaders = []*kernel.BlockHeader{ //nolint:gochecknoglobals // data that is used across all test funcs
	testBlockHeader,
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

var expectedPbBlockHeaders = &pb.BlockHeaders{ //nolint:gochecknoglobals // data that is used across all test funcs
	Headers: []*pb.BlockHeader{
		expectedPbBlockHeader,
		{
			Version:       []byte("v2"),
			PrevBlockHash: []byte("prevhash2"),
			MerkleRoot:    []byte("merkleroot2"),
			Height:        456,
			Timestamp:     1620000000,
			Target:        789,
			Nonce:         101112,
		},
	},
}

func TestSerializeBlock(t *testing.T) {
	p := NewProtobufEncoder()
	data, err := p.SerializeBlock(testBlock)
	require.NoError(t, err)

	var pbBlock pb.Block
	err = proto.Unmarshal(data, &pbBlock)
	require.NoError(t, err)
	// can't use assert.Equal because of the internal proto fields (state can't be stripped)
	assert.True(t, proto.Equal(expectedPbBlock, &pbBlock))
}

func TestDeserializeBlock(t *testing.T) {
	p := NewProtobufEncoder()
	data, err := proto.Marshal(expectedPbBlock)
	require.NoError(t, err)

	block, err := p.DeserializeBlock(data)
	require.NoError(t, err)
	assert.Equal(t, testBlock, *block)
}

func TestSerializeHeader(t *testing.T) {
	p := NewProtobufEncoder()
	data, err := p.SerializeHeader(*testBlock.Header)
	require.NoError(t, err)

	var pbHeader pb.BlockHeader
	err = proto.Unmarshal(data, &pbHeader)
	require.NoError(t, err)

	// can't use assert.Equal because of the internal proto fields (state can't be stripped)
	assert.True(t, proto.Equal(expectedPbBlock.GetHeader(), &pbHeader))
}

func TestDeserializeHeader(t *testing.T) {
	p := NewProtobufEncoder()
	data, err := proto.Marshal(expectedPbBlock.GetHeader())
	require.NoError(t, err)

	header, err := p.DeserializeHeader(data)
	require.NoError(t, err)
	assert.Equal(t, *testBlock.Header, *header)
}

func TestSerializeHeaders(t *testing.T) {
	p := NewProtobufEncoder()
	data, err := p.SerializeHeaders(testBlockHeaders)
	require.NoError(t, err)

	var pbBlockHeaders pb.BlockHeaders
	err = proto.Unmarshal(data, &pbBlockHeaders)
	require.NoError(t, err)

	// can't use assert.Equal because of the internal proto fields (state can't be stripped)
	assert.True(t, proto.Equal(expectedPbBlockHeaders, &pbBlockHeaders))
}

func TestDeserializeHeaders(t *testing.T) {
	p := NewProtobufEncoder()
	data, err := proto.Marshal(expectedPbBlockHeaders)
	require.NoError(t, err)

	headers, err := p.DeserializeHeaders(data)
	require.NoError(t, err)

	assert.ElementsMatch(t, testBlockHeaders, headers)
}

func TestSerializeTransaction(t *testing.T) {
	p := NewProtobufEncoder()
	data, err := p.SerializeTransaction(*testBlock.Transactions[0])
	require.NoError(t, err)

	var pbTransaction pb.Transaction
	err = proto.Unmarshal(data, &pbTransaction)
	require.NoError(t, err)

	expectedPbTransaction := expectedPbBlock.GetTransactions()[0]
	// can't use assert.Equal because of the internal proto fields (state can't be stripped)
	assert.True(t, proto.Equal(expectedPbTransaction, &pbTransaction))
}

func TestDeserializeTransaction(t *testing.T) {
	p := NewProtobufEncoder()
	data, err := proto.Marshal(expectedPbBlock.GetTransactions()[0])
	require.NoError(t, err)

	tx, err := p.DeserializeTransaction(data)
	require.NoError(t, err)
	assert.Equal(t, *testBlock.Transactions[0], *tx)
}

func TestConvertTopbBlock(t *testing.T) {
	expected := expectedPbBlock
	result, err := convertToProtobufBlock(testBlock)
	require.NoError(t, err)

	// can't use assert.Equal because of the internal proto fields (state can't be stripped)
	assert.True(t, proto.Equal(expected, result))
}

func TestConvertFrompbBlock(t *testing.T) {
	expected := testBlock
	result := convertFromProtobufBlock(expectedPbBlock)

	assert.Equal(t, expected, result)
}

func TestConvertTopbBlockHeader(t *testing.T) {
	bh := *testBlock.Header

	expected := &pb.BlockHeader{
		Version:       []byte("v1"),
		PrevBlockHash: []byte("prevhash"),
		MerkleRoot:    []byte("merkleroot"),
		Height:        123,
		Timestamp:     1610000000,
		Target:        456,
		Nonce:         789,
	}
	result := convertToProtobufBlockHeader(bh)

	assert.Equal(t, expected, result)
}

func TestConvertFrompbBlockHeader(t *testing.T) {
	expected := testBlock.Header
	result := convertFromProtobufBlockHeader(expectedPbBlock.GetHeader())

	assert.Equal(t, expected, result)
}

func TestConvertTopbTransaction(t *testing.T) {
	tx := *testBlock.Transactions[0]

	expected := &pb.Transaction{
		Id: []byte("txid1"),
		Vin: []*pb.TxInput{
			{
				Txid:      []byte("txid0"),
				Vout:      0,
				ScriptSig: "sig1",
				PubKey:    "pubkey1",
			},
		},
		Vout: []*pb.TxOutput{
			{
				Amount:       100,
				ScriptPubKey: "scriptpubkey1",
				PubKey:       "pubkey1",
			},
		},
	}
	result := convertToProtobufTransaction(tx)

	assert.Equal(t, expected, result)
}

func TestConvertFrompbTransaction(t *testing.T) {
	tx := *testBlock.Transactions[0]
	expected := tx
	result := convertFromProtobufTransaction(expectedPbBlock.GetTransactions()[0])

	assert.Equal(t, expected, result)
}

func TestConvertTopbTxInput(t *testing.T) {
	input := testBlock.Transactions[0].Vin[0]

	expected := &pb.TxInput{
		Txid:      []byte("txid0"),
		Vout:      0,
		ScriptSig: "sig1",
		PubKey:    "pubkey1",
	}
	result := convertToProtobufTxInput(input)

	assert.Equal(t, expected, result)
}

func TestConvertFrompbTxInput(t *testing.T) {
	expected := testBlock.Transactions[0].Vin[0]
	result := convertFromProtobufTxInput(expectedPbBlock.GetTransactions()[0].GetVin()[0])

	assert.Equal(t, expected, result)
}

func TestConvertTopbTxOutput(t *testing.T) {
	output := testBlock.Transactions[0].Vout[0]

	expected := &pb.TxOutput{
		Amount:       100,
		ScriptPubKey: "scriptpubkey1",
		PubKey:       "pubkey1",
	}
	result := convertToProtobufTxOutput(output)

	assert.Equal(t, expected, result)
}

func TestConvertFrompbTxOutput(t *testing.T) {
	expected := testBlock.Transactions[0].Vout[0]
	result := convertFromProtobufTxOutput(expectedPbBlock.GetTransactions()[0].GetVout()[0])

	assert.Equal(t, expected, result)
}

func TestConvertTopbTxInputs(t *testing.T) {
	inputs := testBlock.Transactions[0].Vin

	expected := []*pb.TxInput{
		{
			Txid:      []byte("txid0"),
			Vout:      0,
			ScriptSig: "sig1",
			PubKey:    "pubkey1",
		},
	}
	result := convertToProtobufTxInputs(inputs)

	assert.Equal(t, expected, result)
}

func TestConvertFrompbTxInputs(t *testing.T) {
	pbInputs := expectedPbBlock.GetTransactions()[0].GetVin()

	expected := testBlock.Transactions[0].Vin
	result := convertFromProtobufTxInputs(pbInputs)

	assert.Equal(t, expected, result)
}

func TestConvertTopbTxOutputs(t *testing.T) {
	outputs := testBlock.Transactions[0].Vout

	expected := []*pb.TxOutput{
		{
			Amount:       100,
			ScriptPubKey: "scriptpubkey1",
			PubKey:       "pubkey1",
		},
	}
	result := convertToProtobufTxOutputs(outputs)

	assert.Equal(t, expected, result)
}

func TestConvertFrompbTxOutputs(t *testing.T) {
	pbOutputs := expectedPbBlock.GetTransactions()[0].GetVout()

	expected := testBlock.Transactions[0].Vout
	result := convertFromProtobufTxOutputs(pbOutputs)

	assert.Equal(t, expected, result)
}

func TestUtf8InvalidCharacters(t *testing.T) {
	ecdsa := sign.NewECDSASignature()
	pubKey, _, err := ecdsa.NewKeyPair()
	require.NoError(t, err)

	p := NewProtobufEncoder()
	block := kernel.Block{
		Header: testBlockHeader,
		Transactions: []*kernel.Transaction{
			kernel.NewCoinbaseTransaction(fmt.Sprintf("%s", pubKey), 50, 0),
		},
		Hash: []byte("blockhash"),
	}

	data, err := p.SerializeBlock(block)
	require.NoError(t, err)

	_, err = p.DeserializeBlock(data)
	require.NoError(t, err)
}

func TestNoNilPointerExceptionsSerialize(t *testing.T) {
	p := NewProtobufEncoder()

	_, err := p.SerializeBlock(kernel.Block{})
	require.Error(t, err)

	_, err = p.SerializeHeader(kernel.BlockHeader{})
	require.NoError(t, err)

	_, err = p.SerializeHeaders([]*kernel.BlockHeader{})
	require.NoError(t, err)

	_, err = p.SerializeHeaders([]*kernel.BlockHeader{{}})
	require.NoError(t, err)

	_, err = p.SerializeTransaction(kernel.Transaction{})
	require.NoError(t, err)
}

func TestNoNilPointerExceptionsDeserialize(t *testing.T) {
	p := NewProtobufEncoder()

	_, err := p.DeserializeBlock([]byte{})
	require.NoError(t, err)

	_, err = p.DeserializeHeader([]byte{})
	require.NoError(t, err)

	_, err = p.DeserializeHeaders([]byte{})
	require.NoError(t, err)

	_, err = p.DeserializeHeaders([]byte{})
	require.NoError(t, err)
}
