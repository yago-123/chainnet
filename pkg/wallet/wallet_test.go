package wallet //nolint:testpackage // don't create separate package for tests

import (
	"chainnet/pkg/consensus/miner"
	"chainnet/pkg/consensus/validator"
	"chainnet/pkg/crypto/hash"
	"chainnet/pkg/kernel"
	"chainnet/pkg/script"
	mockHash "chainnet/tests/mocks/crypto/hash"
	mockSign "chainnet/tests/mocks/crypto/sign"
	"testing"

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

	wallet, err := NewWallet([]byte("0.0.1"), miner.NewProofOfWork(1, hash.NewSHA256()), validator.NewLightValidator(), &signer, &mockHash.MockHashing{})
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

	wallet, err := NewWallet([]byte("0.0.1"), miner.NewProofOfWork(1, hash.NewSHA256()), validator.NewLightValidator(), &signer, &mockHash.MockHashing{})
	require.NoError(t, err)
	// send transaction with correct target and empty tx fee
	tx, err := wallet.SendTransaction("pubkey-1", 10, 0, utxos)
	expectedTx := &kernel.Transaction{
		ID: []byte{0x69, 0x98, 0xc9, 0xa8, 0xea, 0xda, 0xf3, 0x31, 0xd7, 0xac, 0x4e, 0xb0, 0x4a, 0x1c, 0xd8, 0xb4, 0x15, 0x65, 0x51, 0x83, 0x50, 0x4b, 0x79, 0xa4, 0x97, 0xea, 0xa9, 0x9f, 0xd3, 0xb6, 0xc9, 0xb5},
		Vin: []kernel.TxInput{
			kernel.NewInput([]byte("random-id-0"), 1, "Inputs:random-id-01random-id-13random-id-21random-id-38Outputs:10\x00KozLnpdoo68 OP_CHECKSIGpubkey-13\x00KozLnpdoo69 OP_CHECKSIGpubkey-2-signed", "pubkey-2"),
			kernel.NewInput([]byte("random-id-1"), 3, "Inputs:random-id-01random-id-13random-id-21random-id-38Outputs:10\x00KozLnpdoo68 OP_CHECKSIGpubkey-13\x00KozLnpdoo69 OP_CHECKSIGpubkey-2-signed", "pubkey-2"),
			kernel.NewInput([]byte("random-id-2"), 1, "Inputs:random-id-01random-id-13random-id-21random-id-38Outputs:10\x00KozLnpdoo68 OP_CHECKSIGpubkey-13\x00KozLnpdoo69 OP_CHECKSIGpubkey-2-signed", "pubkey-2"),
			kernel.NewInput([]byte("random-id-3"), 8, "Inputs:random-id-01random-id-13random-id-21random-id-38Outputs:10\x00KozLnpdoo68 OP_CHECKSIGpubkey-13\x00KozLnpdoo69 OP_CHECKSIGpubkey-2-signed", "pubkey-2"),
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
		ID: []byte{0xdf, 0xba, 0xa2, 0x2, 0x58, 0xb5, 0x9, 0x28, 0x34, 0x1b, 0x6d, 0x31, 0x91, 0xca, 0xfc, 0x86, 0x23, 0x9e, 0xea, 0x97, 0xb9, 0xc8, 0xa7, 0xb9, 0x20, 0xd7, 0xf3, 0x86, 0x91, 0x1, 0x89, 0xc},
		Vin: []kernel.TxInput{
			kernel.NewInput([]byte("random-id-0"), 1, "Inputs:random-id-01random-id-13random-id-21random-id-38Outputs:10\x00KozLnpdoo6A OP_CHECKSIGpubkey-31\x00KozLnpdoo69 OP_CHECKSIGpubkey-2-signed", "pubkey-2"),
			kernel.NewInput([]byte("random-id-1"), 3, "Inputs:random-id-01random-id-13random-id-21random-id-38Outputs:10\x00KozLnpdoo6A OP_CHECKSIGpubkey-31\x00KozLnpdoo69 OP_CHECKSIGpubkey-2-signed", "pubkey-2"),
			kernel.NewInput([]byte("random-id-2"), 1, "Inputs:random-id-01random-id-13random-id-21random-id-38Outputs:10\x00KozLnpdoo6A OP_CHECKSIGpubkey-31\x00KozLnpdoo69 OP_CHECKSIGpubkey-2-signed", "pubkey-2"),
			kernel.NewInput([]byte("random-id-3"), 8, "Inputs:random-id-01random-id-13random-id-21random-id-38Outputs:10\x00KozLnpdoo6A OP_CHECKSIGpubkey-31\x00KozLnpdoo69 OP_CHECKSIGpubkey-2-signed", "pubkey-2"),
		},
		Vout: []kernel.TxOutput{
			kernel.NewOutput(10, script.P2PK, "pubkey-3"),
			kernel.NewOutput(1, script.P2PK, "pubkey-2"),
		},
	}
	require.NoError(t, err)
	assert.Equal(t, expectedTx2, tx)
}
