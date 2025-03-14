package simplewallet //nolint:testpackage // don't create separate package for tests

import (
	"testing"

	"github.com/yago-123/chainnet/config"
	"github.com/yago-123/chainnet/pkg/consensus/validator"
	"github.com/yago-123/chainnet/pkg/encoding"
	"github.com/yago-123/chainnet/pkg/kernel"
	"github.com/yago-123/chainnet/pkg/script"
	mockHash "github.com/yago-123/chainnet/tests/mocks/crypto/hash"
	mockSign "github.com/yago-123/chainnet/tests/mocks/crypto/sign"

	"github.com/btcsuite/btcutil/base58"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var utxos = []*kernel.UTXO{ //nolint:gochecknoglobals // data that is used across all test funcs
	{TxID: []byte("random-id-0"), OutIdx: 1, Output: kernel.NewOutput(1, script.P2PK, "pubkey-2")},
	{TxID: []byte("random-id-1"), OutIdx: 3, Output: kernel.NewOutput(2, script.P2PK, "pubkey-2")},
	{TxID: []byte("random-id-2"), OutIdx: 1, Output: kernel.NewOutput(5, script.P2PK, "pubkey-2")},
	{TxID: []byte("random-id-3"), OutIdx: 8, Output: kernel.NewOutput(5, script.P2PK, "pubkey-2")},
}

func TestWallet_SendTransaction(t *testing.T) {
	var err error

	signer := mockSign.MockSign{}
	signer.
		On("NewKeyPair").
		Return([]byte("pubkey-2"), []byte("privkey-2"), nil)

	wallet, err := NewWallet(config.NewConfig(), 1, validator.NewLightValidator(&mockHash.FakeHashing{}), &signer, &mockHash.FakeHashing{}, encoding.NewProtobufEncoder())
	require.NoError(t, err)

	// send transaction with a target amount bigger than utxos amount
	_, err = wallet.GenerateNewTransaction(script.P2PK, []byte("pubkey-1"), 100, 1, utxos)
	require.Error(t, err)

	// send transaction with a txFee bigger than utxos amount
	_, err = wallet.GenerateNewTransaction(script.P2PK, []byte("pubkey-1"), 1, 100, utxos)
	require.Error(t, err)

	// send transaction without utxos
	_, err = wallet.GenerateNewTransaction(script.P2PK, []byte("pubkey-1"), 10, 1, []*kernel.UTXO{})
	require.Error(t, err)

	// send transaction with incorrect utxos unlocking scripts
	signer2 := mockSign.MockSign{}
	signer2.
		On("NewKeyPair").
		Return([]byte("pubkey-5"), []byte("privkey-5"), nil)
}

func TestWallet_SendTransactionCheckOutputTx(t *testing.T) {
	var err error

	hasher := &mockHash.FakeHashing{}
	signer := mockSign.MockSign{}
	signer.
		On("NewKeyPair").
		Return([]byte("pubkey-2"), []byte("privkey-2"), nil)

	wallet, err := NewWallet(config.NewConfig(), 1, validator.NewLightValidator(hasher), &signer, hasher, encoding.NewProtobufEncoder())
	require.NoError(t, err)
	// send transaction with correct target and empty tx fee
	tx, err := wallet.GenerateNewTransaction(script.P2PK, []byte("pubkey-1"), 10, 0, utxos)
	expectedTx := &kernel.Transaction{
		ID: tx.ID,
		Vin: []kernel.TxInput{
			kernel.NewInput([]byte("random-id-0"), 1, base58.Encode([]byte("Inputs:random-id-01random-id-13random-id-21random-id-38Outputs:10\x00KozLnpdoo68 OP_CHECKSIGpubkey-13\x00KozLnpdoo69 OP_CHECKSIGpubkey-2-signed")), "pubkey-2"),
			kernel.NewInput([]byte("random-id-1"), 3, base58.Encode([]byte("Inputs:random-id-01random-id-13random-id-21random-id-38Outputs:10\x00KozLnpdoo68 OP_CHECKSIGpubkey-13\x00KozLnpdoo69 OP_CHECKSIGpubkey-2-signed")), "pubkey-2"),
			kernel.NewInput([]byte("random-id-2"), 1, base58.Encode([]byte("Inputs:random-id-01random-id-13random-id-21random-id-38Outputs:10\x00KozLnpdoo68 OP_CHECKSIGpubkey-13\x00KozLnpdoo69 OP_CHECKSIGpubkey-2-signed")), "pubkey-2"),
			kernel.NewInput([]byte("random-id-3"), 8, base58.Encode([]byte("Inputs:random-id-01random-id-13random-id-21random-id-38Outputs:10\x00KozLnpdoo68 OP_CHECKSIGpubkey-13\x00KozLnpdoo69 OP_CHECKSIGpubkey-2-signed")), "pubkey-2"),
		},
		Vout: []kernel.TxOutput{
			kernel.NewOutput(10, script.P2PK, "pubkey-1"),
			kernel.NewOutput(3, script.P2PK, "pubkey-2"),
		},
	}
	require.NoError(t, err)
	assert.Equal(t, expectedTx, tx)

	// send transaction with correct target and some tx fee
	tx, err = wallet.GenerateNewTransaction(script.P2PK, []byte("pubkey-3"), 10, 2, utxos)
	expectedTx2 := &kernel.Transaction{
		ID: tx.ID,
		Vin: []kernel.TxInput{
			kernel.NewInput([]byte("random-id-0"), 1, base58.Encode([]byte("Inputs:random-id-01random-id-13random-id-21random-id-38Outputs:10\x00KozLnpdoo6A OP_CHECKSIGpubkey-31\x00KozLnpdoo69 OP_CHECKSIGpubkey-2-signed")), "pubkey-2"),
			kernel.NewInput([]byte("random-id-1"), 3, base58.Encode([]byte("Inputs:random-id-01random-id-13random-id-21random-id-38Outputs:10\x00KozLnpdoo6A OP_CHECKSIGpubkey-31\x00KozLnpdoo69 OP_CHECKSIGpubkey-2-signed")), "pubkey-2"),
			kernel.NewInput([]byte("random-id-2"), 1, base58.Encode([]byte("Inputs:random-id-01random-id-13random-id-21random-id-38Outputs:10\x00KozLnpdoo6A OP_CHECKSIGpubkey-31\x00KozLnpdoo69 OP_CHECKSIGpubkey-2-signed")), "pubkey-2"),
			kernel.NewInput([]byte("random-id-3"), 8, base58.Encode([]byte("Inputs:random-id-01random-id-13random-id-21random-id-38Outputs:10\x00KozLnpdoo6A OP_CHECKSIGpubkey-31\x00KozLnpdoo69 OP_CHECKSIGpubkey-2-signed")), "pubkey-2"),
		},
		Vout: []kernel.TxOutput{
			kernel.NewOutput(10, script.P2PK, "pubkey-3"),
			kernel.NewOutput(1, script.P2PK, "pubkey-2"),
		},
	}
	require.NoError(t, err)
	assert.Equal(t, expectedTx2, tx)
}
