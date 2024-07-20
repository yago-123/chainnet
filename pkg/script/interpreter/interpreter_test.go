package interpreter //nolint:testpackage // don't create separate package for tests

import (
	"chainnet/pkg/crypto"
	"chainnet/pkg/crypto/hash"
	"chainnet/pkg/crypto/sign"
	"chainnet/pkg/kernel"
	"chainnet/pkg/script"
	mockHash "chainnet/tests/mocks/crypto/hash"
	mockSign "chainnet/tests/mocks/crypto/sign"
	"fmt"
	"testing"

	"github.com/btcsuite/btcutil/base58"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// tx1P2PK contains a single input and a single output
var tx1P2PK = kernel.NewTransaction( //nolint:gochecknoglobals // ignore linter in this case
	[]kernel.TxInput{
		kernel.NewInput([]byte("transaction-1"), 1, "", "pubKey-1"),
	},
	[]kernel.TxOutput{
		kernel.NewOutput(50, script.P2PK, "pubKey-1"),
	},
)

// tx2P2PK contains multiple inputs with same public key
var tx2P2PK = kernel.NewTransaction( //nolint:gochecknoglobals // ignore linter in this case
	[]kernel.TxInput{
		kernel.NewInput([]byte("transaction-1"), 1, "", "pubKey-1"),
		kernel.NewInput([]byte("transaction-2"), 1, "", "pubKey-1"),
	},
	[]kernel.TxOutput{
		kernel.NewOutput(50, script.P2PK, "pubKey-1"),
	},
)

// todo() modify the inputs so it does not match tx2P2PK
// tx3P2PK contains multiple inputs with different public keys
var tx3P2PK = kernel.NewTransaction( //nolint:gochecknoglobals // ignore linter in this case
	[]kernel.TxInput{
		kernel.NewInput([]byte("transaction-1"), 1, "", "pubKey-1"),
		kernel.NewInput([]byte("transaction-2"), 3, "", "pubKey-2"),
		kernel.NewInput([]byte("transaction-3"), 0, "", "pubKey-3"),
	},
	[]kernel.TxOutput{
		kernel.NewOutput(50, script.P2PK, "pubKey-1"),
		kernel.NewOutput(1, script.P2PK, "pubKey-2"),
	},
)

// tx1P2PKH contains a single input and a single output
var tx1P2PKH = kernel.NewTransaction( //nolint:gochecknoglobals // ignore linter in this case
	[]kernel.TxInput{
		kernel.NewInput([]byte("transaction-1"), 1, "", "pubKey-1"),
	},
	[]kernel.TxOutput{
		kernel.NewOutput(50, script.P2PKH, "pubKey-1"),
	},
)

// tx2P2PKH contains multiple inputs with same public key
var tx2P2PKH = kernel.NewTransaction( //nolint:gochecknoglobals // ignore linter in this case
	[]kernel.TxInput{
		kernel.NewInput([]byte("transaction-1"), 1, "", "pubKey-1"),
		kernel.NewInput([]byte("transaction-2"), 1, "", "pubKey-1"),
	},
	[]kernel.TxOutput{
		kernel.NewOutput(50, script.P2PKH, "pubKey-1"),
	},
)

// todo() modify the inputs so it does not match tx2P2PKH
// tx3P2PKH contains multiple inputs with different public keys
var tx3P2PKH = kernel.NewTransaction( //nolint:gochecknoglobals // ignore linter in this case
	[]kernel.TxInput{
		kernel.NewInput([]byte("transaction-1"), 1, "", "pubKey-1"),
		kernel.NewInput([]byte("transaction-2"), 3, "", "pubKey-2"),
		kernel.NewInput([]byte("transaction-3"), 0, "", "pubKey-3"),
	},
	[]kernel.TxOutput{
		kernel.NewOutput(50, script.P2PKH, "pubKey-1"),
		kernel.NewOutput(1, script.P2PKH, "pubKey-2"),
	},
)

func TestRPNInterpreter_GenerateScriptSigWithErrors(t *testing.T) {
	signer := sign.NewECDSASignature()
	interpreter := NewScriptInterpreter(crypto.NewHashedSignature(signer, hash.NewSHA256()))

	pubKey, privKey, err := sign.NewECDSASignature().NewKeyPair()
	require.NoError(t, err)

	// generate the scriptSig with an invalid scriptPubKey
	_, err = interpreter.GenerateScriptSig(
		"invalid script",
		pubKey,
		privKey,
		tx1P2PK,
	)
	require.Error(t, err)

	// generate the scriptSig with an empty private key
	_, err = interpreter.GenerateScriptSig(
		script.NewScript(script.P2PK, pubKey),
		[]byte{},
		[]byte{},
		tx1P2PK,
	)
	require.Error(t, err)

	// generate the scriptSig with an empty transaction
	_, err = interpreter.GenerateScriptSig(
		script.NewScript(script.P2PK, pubKey),
		pubKey,
		privKey,
		&kernel.Transaction{},
	)
	require.Error(t, err)
}

func TestRPNInterpreter_VerifyScriptPubKeyWithErrors(t *testing.T) {
	signer := sign.NewECDSASignature()
	interpreter := NewScriptInterpreter(crypto.NewHashedSignature(signer, hash.NewSHA256()))

	pubKey, privKey, err := signer.NewKeyPair()
	require.NoError(t, err)

	// generate real signature for testing purposes
	realSignature, err := interpreter.GenerateScriptSig(
		script.NewScript(script.P2PK, pubKey),
		pubKey,
		privKey,
		tx1P2PK,
	)
	require.NoError(t, err)

	// check that invalid scripts are not accepted
	valid, err := interpreter.VerifyScriptPubKey(
		"invalid script",
		realSignature,
		tx1P2PK,
	)
	require.Error(t, err)
	require.False(t, valid)

	// check that wrong signatures are accepted but not valid
	valid, err = interpreter.VerifyScriptPubKey(
		script.NewScript(script.P2PK, pubKey),
		"randomsignature",
		tx1P2PK,
	)
	require.NoError(t, err)
	require.False(t, valid)

	// check that empty signatures are not accepted
	valid, err = interpreter.VerifyScriptPubKey(
		script.NewScript(script.P2PK, pubKey),
		"",
		tx1P2PK,
	)
	require.Error(t, err)
	require.False(t, valid)

	// check that empty transactions are not accepted
	valid, err = interpreter.VerifyScriptPubKey(
		script.NewScript(script.P2PK, pubKey),
		realSignature,
		&kernel.Transaction{},
	)
	require.Error(t, err)
	require.False(t, valid)
}

func TestRPNInterpreter_GenerationAndVerificationRealKeysP2PK(t *testing.T) {
	signer := sign.NewECDSASignature()
	interpreter := NewScriptInterpreter(crypto.NewHashedSignature(signer, hash.NewSHA256()))

	pubKey, privKey, err := signer.NewKeyPair()
	require.NoError(t, err)

	// generate the scriptSig to unlock the input
	signature, err := interpreter.GenerateScriptSig(
		script.NewScript(script.P2PK, pubKey),
		pubKey,
		privKey,
		tx1P2PK,
	)
	require.NoError(t, err)

	// check that the scriptSig generated is correct
	valid, err := interpreter.VerifyScriptPubKey(
		script.NewScript(script.P2PK, pubKey),
		signature,
		tx1P2PK,
	)
	require.NoError(t, err)
	assert.True(t, valid)

	// modify the scriptSig and check that is not correct anymore
	modifiedScriptSig := []rune(signature)
	modifiedScriptSig[0] = 'a'
	valid, err = interpreter.VerifyScriptPubKey(
		script.NewScript(script.P2PK, pubKey),
		string(modifiedScriptSig),
		tx1P2PK,
	)
	require.NoError(t, err)
	assert.False(t, valid)
}

func TestRPNInterpreter_GenerationAndVerificationRealKeysP2PKH(t *testing.T) {
	signer := sign.NewECDSASignature()
	interpreter := NewScriptInterpreter(crypto.NewHashedSignature(signer, hash.NewSHA256()))

	pubKey, privKey, err := signer.NewKeyPair()
	require.NoError(t, err)

	// generate the scriptSig to unlock the input
	signature, err := interpreter.GenerateScriptSig(
		script.NewScript(script.P2PKH, pubKey),
		pubKey,
		privKey,
		tx1P2PKH,
	)
	require.NoError(t, err)

	// check that the scriptSig generated is correct
	valid, err := interpreter.VerifyScriptPubKey(
		script.NewScript(script.P2PKH, pubKey),
		signature,
		tx1P2PKH,
	)
	require.NoError(t, err)
	assert.True(t, valid)

	// modify the scriptSig and check that is not correct anymore
	modifiedScriptSig := []rune(signature)
	modifiedScriptSig[0] = 'a'
	valid, err = interpreter.VerifyScriptPubKey(
		script.NewScript(script.P2PKH, pubKey),
		string(modifiedScriptSig),
		tx1P2PKH,
	)
	require.NoError(t, err)
	assert.False(t, valid)
}

func TestRPNInterpreter_GenerateScriptSigP2PKMocked(t *testing.T) {
	interpreter := NewScriptInterpreter(crypto.NewHashedSignature(&mockSign.MockSign{}, &mockHash.MockHashing{}))

	pubKey, privKey, err := sign.NewECDSASignature().NewKeyPair()
	require.NoError(t, err)

	signature, err := interpreter.GenerateScriptSig(
		script.NewScript(script.P2PK, pubKey),
		pubKey,
		privKey,
		tx1P2PK,
	)
	require.NoError(t, err)
	assert.Equal(t, base58.Encode([]byte(fmt.Sprintf("%s-hashed-signed", tx1P2PK.AssembleForSigning()))), signature)

	signature, err = interpreter.GenerateScriptSig(
		script.NewScript(script.P2PK, pubKey),
		pubKey,
		privKey,
		tx2P2PK,
	)
	require.NoError(t, err)
	assert.Equal(t, base58.Encode([]byte(fmt.Sprintf("%s-hashed-signed", tx2P2PK.AssembleForSigning()))), signature)

	signature, err = interpreter.GenerateScriptSig(
		script.NewScript(script.P2PK, pubKey),
		pubKey,
		privKey,
		tx3P2PK,
	)
	require.NoError(t, err)
	assert.Equal(t, base58.Encode([]byte(fmt.Sprintf("%s-hashed-signed", tx3P2PK.AssembleForSigning()))), signature)
}

func TestRPNInterpreter_GenerateScriptSigP2PKHMocked(t *testing.T) {
	interpreter := NewScriptInterpreter(crypto.NewHashedSignature(&mockSign.MockSign{}, &mockHash.MockHashing{}))

	pubKey, privKey, err := sign.NewECDSASignature().NewKeyPair()
	require.NoError(t, err)

	signature, err := interpreter.GenerateScriptSig(
		script.NewScript(script.P2PKH, pubKey),
		pubKey,
		privKey,
		tx1P2PK,
	)
	require.NoError(t, err)
	assert.Equal(t, fmt.Sprintf("%s %s", base58.Encode([]byte(fmt.Sprintf("%s-hashed-signed", tx1P2PK.AssembleForSigning()))), base58.Encode(pubKey)), signature)

	signature, err = interpreter.GenerateScriptSig(
		script.NewScript(script.P2PKH, pubKey),
		pubKey,
		privKey,
		tx2P2PK,
	)
	require.NoError(t, err)
	assert.Equal(t, fmt.Sprintf("%s %s", base58.Encode([]byte(fmt.Sprintf("%s-hashed-signed", tx2P2PK.AssembleForSigning()))), base58.Encode(pubKey)), signature)

	signature, err = interpreter.GenerateScriptSig(
		script.NewScript(script.P2PKH, pubKey),
		pubKey,
		privKey,
		tx3P2PK,
	)
	require.NoError(t, err)
	assert.Equal(t, fmt.Sprintf("%s %s", base58.Encode([]byte(fmt.Sprintf("%s-hashed-signed", tx3P2PK.AssembleForSigning()))), base58.Encode(pubKey)), signature)
}

func TestRPNInterpreter_VerifyScriptPubKeyP2PKMocked(t *testing.T) {
	interpreter := NewScriptInterpreter(crypto.NewHashedSignature(&mockSign.MockSign{}, &mockHash.MockHashing{}))

	pubKey, _, err := sign.NewECDSASignature().NewKeyPair()
	require.NoError(t, err)

	valid, err := interpreter.VerifyScriptPubKey(
		script.NewScript(script.P2PK, pubKey),
		base58.Encode([]byte(fmt.Sprintf("%s-hashed-signed", tx1P2PK.AssembleForSigning()))),
		tx1P2PK,
	)
	require.NoError(t, err)
	assert.True(t, valid)

	valid, err = interpreter.VerifyScriptPubKey(
		script.NewScript(script.P2PK, pubKey),
		base58.Encode([]byte(fmt.Sprintf("%s-hashed-signed", tx2P2PK.AssembleForSigning()))),
		tx2P2PK,
	)
	require.NoError(t, err)
	assert.True(t, valid)

	valid, err = interpreter.VerifyScriptPubKey(
		script.NewScript(script.P2PK, pubKey),
		base58.Encode([]byte(fmt.Sprintf("%s-hashed-signed", tx3P2PK.AssembleForSigning()))),
		tx3P2PK,
	)
	require.NoError(t, err)
	assert.True(t, valid)
}

func TestRPNInterpreter_VerifyScriptPubKeyP2PKHMocked(t *testing.T) {
	interpreter := NewScriptInterpreter(crypto.NewHashedSignature(&mockSign.MockSign{}, &mockHash.MockHashing{}))

	pubKey, _, err := sign.NewECDSASignature().NewKeyPair()
	require.NoError(t, err)

	valid, err := interpreter.VerifyScriptPubKey(
		script.NewScript(script.P2PKH, pubKey),
		fmt.Sprintf("%s %s", base58.Encode([]byte(fmt.Sprintf("%s-hashed-signed", tx1P2PKH.AssembleForSigning()))), base58.Encode(pubKey)),
		tx1P2PKH,
	)
	require.NoError(t, err)
	assert.True(t, valid)

	valid, err = interpreter.VerifyScriptPubKey(
		script.NewScript(script.P2PKH, pubKey),
		fmt.Sprintf("%s %s", base58.Encode([]byte(fmt.Sprintf("%s-hashed-signed", tx2P2PKH.AssembleForSigning()))), base58.Encode(pubKey)),
		tx2P2PKH,
	)
	require.NoError(t, err)
	assert.True(t, valid)

	valid, err = interpreter.VerifyScriptPubKey(
		script.NewScript(script.P2PKH, pubKey),
		fmt.Sprintf("%s %s", base58.Encode([]byte(fmt.Sprintf("%s-hashed-signed", tx3P2PKH.AssembleForSigning()))), base58.Encode(pubKey)),
		tx3P2PKH,
	)
	require.NoError(t, err)
	assert.True(t, valid)

	// verify scriptSig with different pubKey than expected
	valid, err = interpreter.VerifyScriptPubKey(
		script.NewScript(script.P2PKH, pubKey),
		fmt.Sprintf("%s %s", base58.Encode([]byte(fmt.Sprintf("%s-hashed-signed", tx3P2PKH.AssembleForSigning()))), base58.Encode([]byte("differentpubkey"))),
		tx1P2PKH,
	)
	require.Error(t, err)
	assert.False(t, valid)
}
