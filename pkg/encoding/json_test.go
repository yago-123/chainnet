package encoding

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yago-123/chainnet/pkg/kernel"
)

func TestJSONSerializeDeserializeTransaction(t *testing.T) {
	encoder := NewJSONEncoder()
	tx := kernel.NewTransaction(
		[]kernel.TxInput{kernel.NewInput([]byte("tx-id"), 1, "script-sig", "pub-key")},
		[]kernel.TxOutput{{Amount: 10, ScriptPubKey: "script", PubKey: "pub-key"}},
	)
	tx.SetID([]byte("tx-id"))

	data, err := encoder.SerializeTransaction(*tx)
	require.NoError(t, err)
	assert.JSONEq(t, `{
		"id":"74782d6964",
		"vin":[{"txid":"74782d6964","vout":1,"script_sig":"script-sig","pub_key":"pub-key"}],
		"vout":[{"amount":10,"script_pub_key":"script","pub_key":"pub-key"}]
	}`, string(data))

	decoded, err := encoder.DeserializeTransaction(data)
	require.NoError(t, err)
	assert.Equal(t, tx, decoded)
}

func TestJSONSerializeDeserializeUTXOs(t *testing.T) {
	encoder := NewJSONEncoder()
	utxos := []*kernel.UTXO{
		{
			TxID:   []byte("tx-id"),
			OutIdx: 1,
			Output: kernel.TxOutput{
				Amount:       10,
				ScriptPubKey: "script",
				PubKey:       "pub-key",
			},
		},
	}

	data, err := encoder.SerializeUTXOs(utxos)
	require.NoError(t, err)

	decoded, err := encoder.DeserializeUTXOs(data)
	require.NoError(t, err)
	assert.Equal(t, utxos, decoded)
}

func TestJSONSerializeDeserializeBlock(t *testing.T) {
	encoder := NewJSONEncoder()
	tx := kernel.NewTransaction([]kernel.TxInput{}, []kernel.TxOutput{{Amount: 10, ScriptPubKey: "script", PubKey: "pub-key"}})
	tx.SetID([]byte("tx-id"))
	header := kernel.NewBlockHeader([]byte("v1"), 123, []byte("merkle-root"), 7, []byte("prev-hash"), 1, 42)
	block := kernel.NewBlock(header, []*kernel.Transaction{tx}, []byte("block-hash"))

	data, err := encoder.SerializeBlock(*block)
	require.NoError(t, err)

	decoded, err := encoder.DeserializeBlock(data)
	require.NoError(t, err)
	assert.Equal(t, block, decoded)
}

func TestJSONSerializeDeserializeBool(t *testing.T) {
	encoder := NewJSONEncoder()

	data, err := encoder.SerializeBool(true)
	require.NoError(t, err)
	assert.JSONEq(t, `true`, string(data))

	decoded, err := encoder.DeserializeBool(data)
	require.NoError(t, err)
	assert.True(t, decoded)
}
