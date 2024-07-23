package wallet //nolint:testpackage // don't create separate package for tests

import (
	"chainnet/pkg/consensus/validator"
	"chainnet/pkg/kernel"
	"chainnet/pkg/script"
	mockHash "chainnet/tests/mocks/crypto/hash"
	mockSign "chainnet/tests/mocks/crypto/sign"
	"testing"

	"github.com/btcsuite/btcutil/base58"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var utxos = []*kernel.UnspentOutput{ //nolint:gochecknoglobals // data that is used across all test funcs
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

	wallet, err := NewWallet([]byte("0.0.1"), validator.NewLightValidator(), &signer, &mockHash.MockHashing{})
	require.NoError(t, err)

	// send transaction with a target amount bigger than utxos amount
	_, err = wallet.SendTransaction("pubkey-1", 100, 1, utxos)
	require.Error(t, err)

	// send transaction with a txFee bigger than utxos amount
	_, err = wallet.SendTransaction("pubkey-1", 1, 100, utxos)
	require.Error(t, err)

	// send transaction without utxos
	_, err = wallet.SendTransaction("pubkey-1", 10, 1, []*kernel.UnspentOutput{})
	require.Error(t, err)

	// send transaction with incorrect utxos unlocking scripts
	signer2 := mockSign.MockSign{}
	signer2.
		On("NewKeyPair").
		Return([]byte("pubkey-5"), []byte("privkey-5"), nil)
	// wallet2, err := NewWallet([]byte("0.0.1"), miner.NewProofOfWork(1, hash.NewSHA256()), validator.NewLightValidator(), &signer2, &mockHash.MockHashing{})
	// require.NoError(t, err)

	// todo(): add script signature validator? probably depends on type of wallet: nespv, spv, full node wallet...
	// _, err = wallet2.SendTransaction("pubkey-1", 10, 1, utxos)
	// require.Error(t, err)
}

func TestWallet_SendTransactionCheckOutputTx(t *testing.T) {
	var err error

	signer := mockSign.MockSign{}
	signer.
		On("NewKeyPair").
		Return([]byte("pubkey-2"), []byte("privkey-2"), nil)

	wallet, err := NewWallet([]byte("0.0.1"), validator.NewLightValidator(), &signer, &mockHash.MockHashing{})
	require.NoError(t, err)
	// send transaction with correct target and empty tx fee
	tx, err := wallet.SendTransaction("pubkey-1", 10, 0, utxos)
	expectedTx := &kernel.Transaction{
		ID: []byte{0x49, 0x6e, 0x70, 0x75, 0x74, 0x73, 0x3a, 0x72, 0x61, 0x6e, 0x64, 0x6f, 0x6d, 0x2d, 0x69, 0x64, 0x2d, 0x30, 0x31, 0x70, 0x75, 0x62, 0x6b, 0x65, 0x79, 0x2d, 0x32, 0x72, 0x61, 0x6e, 0x64, 0x6f, 0x6d, 0x2d, 0x69, 0x64, 0x2d, 0x31, 0x33, 0x70, 0x75, 0x62, 0x6b, 0x65, 0x79, 0x2d, 0x32, 0x72, 0x61, 0x6e, 0x64, 0x6f, 0x6d, 0x2d, 0x69, 0x64, 0x2d, 0x32, 0x31, 0x70, 0x75, 0x62, 0x6b, 0x65, 0x79, 0x2d, 0x32, 0x72, 0x61, 0x6e, 0x64, 0x6f, 0x6d, 0x2d, 0x69, 0x64, 0x2d, 0x33, 0x38, 0x70, 0x75, 0x62, 0x6b, 0x65, 0x79, 0x2d, 0x32, 0x4f, 0x75, 0x74, 0x70, 0x75, 0x74, 0x73, 0x3a, 0x31, 0x30, 0x0, 0x4b, 0x6f, 0x7a, 0x4c, 0x6e, 0x70, 0x64, 0x6f, 0x6f, 0x36, 0x38, 0x20, 0x4f, 0x50, 0x5f, 0x43, 0x48, 0x45, 0x43, 0x4b, 0x53, 0x49, 0x47, 0x70, 0x75, 0x62, 0x6b, 0x65, 0x79, 0x2d, 0x31, 0x33, 0x0, 0x4b, 0x6f, 0x7a, 0x4c, 0x6e, 0x70, 0x64, 0x6f, 0x6f, 0x36, 0x39, 0x20, 0x4f, 0x50, 0x5f, 0x43, 0x48, 0x45, 0x43, 0x4b, 0x53, 0x49, 0x47, 0x70, 0x75, 0x62, 0x6b, 0x65, 0x79, 0x2d, 0x32, 0x2d, 0x68, 0x61, 0x73, 0x68, 0x65, 0x64},
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
	tx, err = wallet.SendTransaction("pubkey-3", 10, 2, utxos)
	expectedTx2 := &kernel.Transaction{
		ID: []byte{0x49, 0x6e, 0x70, 0x75, 0x74, 0x73, 0x3a, 0x72, 0x61, 0x6e, 0x64, 0x6f, 0x6d, 0x2d, 0x69, 0x64, 0x2d, 0x30, 0x31, 0x70, 0x75, 0x62, 0x6b, 0x65, 0x79, 0x2d, 0x32, 0x72, 0x61, 0x6e, 0x64, 0x6f, 0x6d, 0x2d, 0x69, 0x64, 0x2d, 0x31, 0x33, 0x70, 0x75, 0x62, 0x6b, 0x65, 0x79, 0x2d, 0x32, 0x72, 0x61, 0x6e, 0x64, 0x6f, 0x6d, 0x2d, 0x69, 0x64, 0x2d, 0x32, 0x31, 0x70, 0x75, 0x62, 0x6b, 0x65, 0x79, 0x2d, 0x32, 0x72, 0x61, 0x6e, 0x64, 0x6f, 0x6d, 0x2d, 0x69, 0x64, 0x2d, 0x33, 0x38, 0x70, 0x75, 0x62, 0x6b, 0x65, 0x79, 0x2d, 0x32, 0x4f, 0x75, 0x74, 0x70, 0x75, 0x74, 0x73, 0x3a, 0x31, 0x30, 0x0, 0x4b, 0x6f, 0x7a, 0x4c, 0x6e, 0x70, 0x64, 0x6f, 0x6f, 0x36, 0x41, 0x20, 0x4f, 0x50, 0x5f, 0x43, 0x48, 0x45, 0x43, 0x4b, 0x53, 0x49, 0x47, 0x70, 0x75, 0x62, 0x6b, 0x65, 0x79, 0x2d, 0x33, 0x31, 0x0, 0x4b, 0x6f, 0x7a, 0x4c, 0x6e, 0x70, 0x64, 0x6f, 0x6f, 0x36, 0x39, 0x20, 0x4f, 0x50, 0x5f, 0x43, 0x48, 0x45, 0x43, 0x4b, 0x53, 0x49, 0x47, 0x70, 0x75, 0x62, 0x6b, 0x65, 0x79, 0x2d, 0x32, 0x2d, 0x68, 0x61, 0x73, 0x68, 0x65, 0x64},
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
