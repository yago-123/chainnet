package encoding //nolint:testpackage // don't create separate package for tests

import (
	"testing"

	"github.com/yago-123/chainnet/pkg/crypto/sign"
	"github.com/yago-123/chainnet/pkg/kernel"
	pb "github.com/yago-123/chainnet/pkg/p2p/protobuf"

	"github.com/stretchr/testify/require"

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
					PubKey:    "7075626b657931", // hexadecimal encoded to prevent UTF-8 issues
				},
			},
			Vout: []*pb.TxOutput{
				{
					Amount:       100,
					ScriptPubKey: "scriptpubkey1",
					PubKey:       "7075626b657931", // hexadecimal encoded to prevent UTF-8 issues
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

func TestSerializeTransactions(t *testing.T) {
	// Reuse testBlock.Transactions for the test data
	p := NewProtobufEncoder()
	data, err := p.SerializeTransactions(testBlock.Transactions)
	require.NoError(t, err)

	expectedPbTransactions := []*pb.Transaction{
		{
			Id: []byte("txid1"),
			Vin: []*pb.TxInput{
				{
					Txid:      []byte("txid0"),
					Vout:      0,
					ScriptSig: "sig1",
					PubKey:    "7075626b657931", // Hex encoded to avoid UTF-8 issues
				},
			},
			Vout: []*pb.TxOutput{
				{
					Amount:       100,
					ScriptPubKey: "scriptpubkey1",
					PubKey:       "7075626b657931", // Hex encoded to avoid UTF-8 issues
				},
			},
		},
	}

	var pbTransactions pb.Transactions
	err = proto.Unmarshal(data, &pbTransactions)
	require.NoError(t, err)

	// Verify the equality using proto.Equal to account for proto internal fields
	for i := range pbTransactions.Transactions {
		assert.True(t, proto.Equal(expectedPbTransactions[i], pbTransactions.Transactions[i]))
	}
}

func TestDeserializeTransactions(t *testing.T) {
	p := NewProtobufEncoder()

	// Create the protobuf representation of transactions and marshal it
	expectedPbTransactions := &pb.Transactions{
		Transactions: []*pb.Transaction{
			{
				Id: []byte("txid1"),
				Vin: []*pb.TxInput{
					{
						Txid:      []byte("txid0"),
						Vout:      0,
						ScriptSig: "sig1",
						PubKey:    "7075626b657931",
					},
				},
				Vout: []*pb.TxOutput{
					{
						Amount:       100,
						ScriptPubKey: "scriptpubkey1",
						PubKey:       "7075626b657931",
					},
				},
			},
		},
	}

	data, err := proto.Marshal(expectedPbTransactions)
	require.NoError(t, err)

	// Deserialize back into kernel.Transaction struct
	transactions, err := p.DeserializeTransactions(data)
	require.NoError(t, err)

	// Assert that the deserialized transactions match the testBlock transactions
	assert.ElementsMatch(t, testBlock.Transactions, transactions)
}

func TestSerializeUTXO(t *testing.T) {
	// Example UTXO
	utxo := kernel.UTXO{
		TxID:   []byte("utxoid1"),
		OutIdx: 0,
		Output: kernel.TxOutput{
			Amount:       50,
			ScriptPubKey: "scriptpubkey1",
			PubKey:       "pubkey1",
		},
	}

	// Serialize the UTXO
	p := NewProtobufEncoder()
	data, err := p.SerializeUTXO(utxo)
	require.NoError(t, err)

	// Expected protobuf representation of the UTXO
	expectedPbUTXO := &pb.UTXO{
		Txid: []byte("utxoid1"),
		Vout: 0,
		Output: &pb.TxOutput{
			Amount:       50,
			ScriptPubKey: "scriptpubkey1",
			PubKey:       "7075626b657931", // Hex encoded PubKey
		},
	}

	var pbUtxo pb.UTXO
	err = proto.Unmarshal(data, &pbUtxo)
	require.NoError(t, err)

	// Verify the equality using proto.Equal
	assert.True(t, proto.Equal(expectedPbUTXO, &pbUtxo))
}

func TestDeserializeUTXO(t *testing.T) {
	p := NewProtobufEncoder()

	// Expected protobuf representation of the UTXO
	expectedPbUTXO := &pb.UTXO{
		Txid: []byte("utxoid1"),
		Vout: 0,
		Output: &pb.TxOutput{
			Amount:       50,
			ScriptPubKey: "scriptpubkey1",
			PubKey:       "7075626b657931", // Hex encoded PubKey
		},
	}

	// Serialize the protobuf UTXO for testing deserialization
	data, err := proto.Marshal(expectedPbUTXO)
	require.NoError(t, err)

	// Deserialize the protobuf data back to a kernel.UTXO
	utxo, err := p.DeserializeUTXO(data)
	require.NoError(t, err)

	// Expected UTXO
	expectedUTXO := &kernel.UTXO{
		TxID:   []byte("utxoid1"),
		OutIdx: 0,
		Output: kernel.TxOutput{
			Amount:       50,
			ScriptPubKey: "scriptpubkey1",
			PubKey:       "pubkey1",
		},
	}

	// Assert that the deserialized UTXO matches the original UTXO
	assert.Equal(t, expectedUTXO, utxo)
}

func TestSerializeUTXOs(t *testing.T) {
	// Create a list of UTXOs for testing
	utxos := []*kernel.UTXO{
		{
			TxID:   []byte("utxoid1"),
			OutIdx: 0,
			Output: kernel.TxOutput{
				Amount:       50,
				ScriptPubKey: "scriptpubkey1",
				PubKey:       "pubkey1",
			},
		},
		{
			TxID:   []byte("utxoid2"),
			OutIdx: 1,
			Output: kernel.TxOutput{
				Amount:       100,
				ScriptPubKey: "scriptpubkey2",
				PubKey:       "pubkey2",
			},
		},
	}

	// Serialize the list of UTXOs
	p := NewProtobufEncoder()
	data, err := p.SerializeUTXOs(utxos)
	require.NoError(t, err)

	// Expected protobuf representation of UTXOs
	expectedPbUTXOs := &pb.UTXOs{
		Utxos: []*pb.UTXO{
			{
				Txid: []byte("utxoid1"),
				Vout: 0,
				Output: &pb.TxOutput{
					Amount:       50,
					ScriptPubKey: "scriptpubkey1",
					PubKey:       "7075626b657931", // Hex encoded PubKey
				},
			},
			{
				Txid: []byte("utxoid2"),
				Vout: 1,
				Output: &pb.TxOutput{
					Amount:       100,
					ScriptPubKey: "scriptpubkey2",
					PubKey:       "7075626b657932", // Hex encoded PubKey
				},
			},
		},
	}

	var pbUtxos pb.UTXOs
	err = proto.Unmarshal(data, &pbUtxos)
	require.NoError(t, err)

	// Verify the equality using proto.Equal
	assert.True(t, proto.Equal(expectedPbUTXOs, &pbUtxos))
}

func TestDeserializeUTXOs(t *testing.T) {
	p := NewProtobufEncoder()

	// Expected protobuf representation of UTXOs
	expectedPbUTXOs := &pb.UTXOs{
		Utxos: []*pb.UTXO{
			{
				Txid: []byte("utxoid1"),
				Vout: 0,
				Output: &pb.TxOutput{
					Amount:       50,
					ScriptPubKey: "scriptpubkey1",
					PubKey:       "7075626b657931", // Hex encoded PubKey
				},
			},
			{
				Txid: []byte("utxoid2"),
				Vout: 1,
				Output: &pb.TxOutput{
					Amount:       100,
					ScriptPubKey: "scriptpubkey2",
					PubKey:       "7075626b657932", // Hex encoded PubKey
				},
			},
		},
	}

	// Serialize the protobuf UTXOs for testing deserialization
	data, err := proto.Marshal(expectedPbUTXOs)
	require.NoError(t, err)

	// Deserialize the protobuf data back to a list of kernel.UTXO
	utxos, err := p.DeserializeUTXOs(data)
	require.NoError(t, err)

	// Expected UTXO list
	expectedUTXOs := []*kernel.UTXO{
		{
			TxID:   []byte("utxoid1"),
			OutIdx: 0,
			Output: kernel.TxOutput{
				Amount:       50,
				ScriptPubKey: "scriptpubkey1",
				PubKey:       "pubkey1",
			},
		},
		{
			TxID:   []byte("utxoid2"),
			OutIdx: 1,
			Output: kernel.TxOutput{
				Amount:       100,
				ScriptPubKey: "scriptpubkey2",
				PubKey:       "pubkey2",
			},
		},
	}

	// Assert that the deserialized UTXOs match the original UTXO list
	assert.Equal(t, expectedUTXOs, utxos)
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
	result, err := convertFromProtobufBlock(expectedPbBlock)
	require.NoError(t, err)

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
				PubKey:    "7075626b657931",
			},
		},
		Vout: []*pb.TxOutput{
			{
				Amount:       100,
				ScriptPubKey: "scriptpubkey1",
				PubKey:       "7075626b657931",
			},
		},
	}
	result := convertToProtobufTransaction(tx)

	assert.Equal(t, expected, result)
}

func TestConvertFrompbTransaction(t *testing.T) {
	tx := *testBlock.Transactions[0]
	expected := tx
	result, err := convertFromProtobufTransaction(expectedPbBlock.GetTransactions()[0])
	require.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestConvertTopbTxInput(t *testing.T) {
	input := testBlock.Transactions[0].Vin[0]

	expected := &pb.TxInput{
		Txid:      []byte("txid0"),
		Vout:      0,
		ScriptSig: "sig1",
		PubKey:    "7075626b657931",
	}
	result := convertToProtobufTxInput(input)

	assert.Equal(t, expected, result)
}

func TestConvertFrompbTxInput(t *testing.T) {
	expected := testBlock.Transactions[0].Vin[0]
	result, err := convertFromProtobufTxInput(expectedPbBlock.GetTransactions()[0].GetVin()[0])
	require.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestConvertTopbTxOutput(t *testing.T) {
	output := testBlock.Transactions[0].Vout[0]

	expected := &pb.TxOutput{
		Amount:       100,
		ScriptPubKey: "scriptpubkey1",
		PubKey:       "7075626b657931",
	}
	result := convertToProtobufTxOutput(output)

	assert.Equal(t, expected, result)
}

func TestConvertFrompbTxOutput(t *testing.T) {
	expected := testBlock.Transactions[0].Vout[0]
	result, err := convertFromProtobufTxOutput(expectedPbBlock.GetTransactions()[0].GetVout()[0])
	require.NoError(t, err)

	assert.Equal(t, expected, result)
}

func TestConvertTopbTxInputs(t *testing.T) {
	inputs := testBlock.Transactions[0].Vin

	expected := []*pb.TxInput{
		{
			Txid:      []byte("txid0"),
			Vout:      0,
			ScriptSig: "sig1",
			PubKey:    "7075626b657931",
		},
	}
	result := convertToProtobufTxInputs(inputs)

	assert.Equal(t, expected, result)
}

func TestConvertFrompbTxInputs(t *testing.T) {
	pbInputs := expectedPbBlock.GetTransactions()[0].GetVin()

	expected := testBlock.Transactions[0].Vin
	result, err := convertFromProtobufTxInputs(pbInputs)
	require.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestConvertTopbTxOutputs(t *testing.T) {
	outputs := testBlock.Transactions[0].Vout

	expected := []*pb.TxOutput{
		{
			Amount:       100,
			ScriptPubKey: "scriptpubkey1",
			PubKey:       "7075626b657931",
		},
	}
	result := convertToProtobufTxOutputs(outputs)

	assert.Equal(t, expected, result)
}

func TestConvertFrompbTxOutputs(t *testing.T) {
	pbOutputs := expectedPbBlock.GetTransactions()[0].GetVout()

	expected := testBlock.Transactions[0].Vout
	result, err := convertFromProtobufTxOutputs(pbOutputs)
	require.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestConvertToProtobufUTXO(t *testing.T) {
	// Create a sample UTXO object
	utxo := kernel.UTXO{
		TxID:   []byte("sampletxid"),
		OutIdx: 0,
		Output: kernel.TxOutput{
			Amount:       50,
			ScriptPubKey: "sampleScriptPubKey",
			PubKey:       "pubkey1", // hex representation
		},
	}

	// Call the function to convert to protobuf UTXO
	pbUTXO := convertToProtobufUTXO(utxo)

	// Expected protobuf UTXO
	expectedPbUTXO := &pb.UTXO{
		Txid: []byte("sampletxid"),
		Vout: 0,
		Output: &pb.TxOutput{
			Amount:       50,
			ScriptPubKey: "sampleScriptPubKey",
			PubKey:       "7075626b657931", // hex encoded
		},
	}

	// Verify the converted protobuf UTXO matches the expected value
	assert.True(t, proto.Equal(expectedPbUTXO, pbUTXO))
}

func TestConvertFromProtobufUTXO(t *testing.T) {
	// Create a sample protobuf UTXO object
	pbUTXO := &pb.UTXO{
		Txid: []byte("sampletxid"),
		Vout: 0,
		Output: &pb.TxOutput{
			Amount:       50,
			ScriptPubKey: "sampleScriptPubKey",
			PubKey:       "7075626b657931", // hex encoded
		},
	}

	// Call the function to convert from protobuf UTXO
	utxo, err := convertFromProtobufUTXO(pbUTXO)
	require.NoError(t, err)

	// Expected kernel UTXO
	expectedUTXO := kernel.UTXO{
		TxID:   []byte("sampletxid"),
		OutIdx: 0,
		Output: kernel.TxOutput{
			Amount:       50,
			ScriptPubKey: "sampleScriptPubKey",
			PubKey:       "pubkey1", // decoded hex representation
		},
	}

	// Verify the converted kernel UTXO matches the expected value
	assert.Equal(t, expectedUTXO, utxo)
}

func TestConvertToAndFromProtobufUTXO(t *testing.T) {
	// Create a sample UTXO object
	utxo := kernel.UTXO{
		TxID:   []byte("sampletxid"),
		OutIdx: 1,
		Output: kernel.TxOutput{
			Amount:       100,
			ScriptPubKey: "anotherScriptPubKey",
			PubKey:       "7075626b657932", // hex representation
		},
	}

	// Convert to protobuf and back to kernel UTXO
	pbUTXO := convertToProtobufUTXO(utxo)
	convertedUTXO, err := convertFromProtobufUTXO(pbUTXO)
	require.NoError(t, err)

	// Verify that the original UTXO matches the UTXO after conversion
	assert.Equal(t, utxo, convertedUTXO)
}

func TestUtf8InvalidCharacters(t *testing.T) {
	ecdsa := sign.NewECDSASignature()
	pubKey, _, err := ecdsa.NewKeyPair()
	require.NoError(t, err)

	p := NewProtobufEncoder()
	block := kernel.Block{
		Header: testBlockHeader,
		Transactions: []*kernel.Transaction{
			kernel.NewCoinbaseTransaction(string(pubKey), 50, 0),
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

	_, err = p.SerializeTransactions([]*kernel.Transaction{})
	require.NoError(t, err)

	_, err = p.SerializeUTXO(kernel.UTXO{})
	require.NoError(t, err)

	_, err = p.SerializeUTXOs([]*kernel.UTXO{})
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

	_, err = p.DeserializeTransaction([]byte{})
	require.NoError(t, err)

	_, err = p.DeserializeTransactions([]byte{})
	require.NoError(t, err)

	_, err = p.DeserializeUTXO([]byte{})
	require.NoError(t, err)

	_, err = p.DeserializeUTXOs([]byte{})
	require.NoError(t, err)
}
