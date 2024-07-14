package wallet //nolint:testpackage // don't create separate package for tests

import (
	"chainnet/pkg/consensus"
	"chainnet/pkg/kernel"
	"chainnet/pkg/script"
	mockHash "chainnet/tests/mocks/crypto/hash"
	mockSign "chainnet/tests/mocks/crypto/sign"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestWallet_SendTransaction(t *testing.T) {
	var err error

	signer := mockSign.MockSign{}
	signer.
		On("NewKeyPair").
		Return([]byte("pubkey-2"), []byte("privkey-2"), nil)
	wallet, err := NewWallet([]byte("0.0.1"), consensus.NewLightValidator(), &signer, &mockHash.MockHashing{})
	require.NoError(t, err)

	utxos := []*kernel.UnspentOutput{
		{TxID: []byte("random-id-0"), OutIdx: 1, Output: kernel.NewOutput(1, script.P2PK, "pubkey-2")},
		{TxID: []byte("random-id-1"), OutIdx: 3, Output: kernel.NewOutput(2, script.P2PK, "pubkey-2")},
		{TxID: []byte("random-id-2"), OutIdx: 1, Output: kernel.NewOutput(5, script.P2PK, "pubkey-2")},
		{TxID: []byte("random-id-3"), OutIdx: 8, Output: kernel.NewOutput(5, script.P2PK, "pubkey-2")},
	}

	// send transaction to an empty pub key
	_, err = wallet.SendTransaction("", 10, 1, utxos)
	require.Error(t, err)

	// send transaction with a target amount bigger than utxos amount
	_, err = wallet.SendTransaction("pubkey-1", 100, 1, utxos)
	require.Error(t, err)

	// send transaction without utxos
	_, err = wallet.SendTransaction("pubkey-1", 10, 1, []*kernel.UnspentOutput{})
	require.Error(t, err)

	// send transaction with correct target and empty tx fee
	tx, err := wallet.SendTransaction("pubkey-1", 10, 0, utxos)
	expectedTx := &kernel.Transaction{
		ID: []byte("aa"),
		Vin: []kernel.TxInput{
			kernel.NewInput([]byte("random-id-0"), 1, "", "pubkey-1"),
			kernel.NewInput([]byte("random-id-1"), 3, "", "pubkey-1"),
			kernel.NewInput([]byte("random-id-2"), 1, "", "pubkey-1"),
			kernel.NewInput([]byte("random-id-3"), 8, "", "pubkey-1"),
		},
		Vout: []kernel.TxOutput{
			kernel.NewOutput(10, script.P2PK, "pubkey-1"),
			kernel.NewOutput(3, script.P2PK, "pubkey-2"),
		},
	}

	require.NoError(t, err)
	assert.Equal(t, expectedTx, tx)

	// send transaction with correct target and some tx fee
	tx, err = wallet.SendTransaction("pubkey-1", 10, 2, utxos)
	require.NoError(t, err)
}
