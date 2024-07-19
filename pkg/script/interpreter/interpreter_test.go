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
	"github.com/btcsuite/btcutil/base58"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// tx1 contains a single input and a single output
var tx1 = kernel.NewTransaction( //nolint:gochecknoglobals // ignore linter in this case
	[]kernel.TxInput{
		kernel.NewInput([]byte("transaction-1"), 1, "", "pubKey-1"),
	},
	[]kernel.TxOutput{
		kernel.NewOutput(50, script.P2PK, "pubKey-1"),
	},
)

// tx2 contains multiple inputs with same public key
var tx2 = kernel.NewTransaction( //nolint:gochecknoglobals // ignore linter in this case
	[]kernel.TxInput{
		kernel.NewInput([]byte("transaction-1"), 1, "", "pubKey-1"),
		kernel.NewInput([]byte("transaction-2"), 1, "", "pubKey-1"),
	},
	[]kernel.TxOutput{
		kernel.NewOutput(50, script.P2PK, "pubKey-1"),
	},
)

// tx3 contains multiple inputs with different public keys
var tx3 = kernel.NewTransaction( //nolint:gochecknoglobals // ignore linter in this case
	[]kernel.TxInput{
		kernel.NewInput([]byte("transaction-1"), 1, "", "pubKey-1"),
		kernel.NewInput([]byte("transaction-2"), 1, "", "pubKey-2"),
	},
	[]kernel.TxOutput{
		kernel.NewOutput(50, script.P2PK, "pubKey-1"),
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
		privKey,
		tx1,
	)
	require.Error(t, err)

	// generate the scriptSig with an empty private key
	_, err = interpreter.GenerateScriptSig(
		script.NewScript(script.P2PK, pubKey),
		[]byte{},
		tx1,
	)
	require.Error(t, err)

	// generate the scriptSig with an empty transaction
	_, err = interpreter.GenerateScriptSig(
		script.NewScript(script.P2PK, pubKey),
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
		privKey,
		tx1,
	)
	require.NoError(t, err)

	// check that invalid scripts are not accepted
	valid, err := interpreter.VerifyScriptPubKey(
		"invalid script",
		realSignature,
		tx1,
	)
	require.Error(t, err)
	require.False(t, valid)

	// check that wrong signatures are accepted but not valid
	valid, err = interpreter.VerifyScriptPubKey(
		script.NewScript(script.P2PK, pubKey),
		"randomsignature",
		tx1,
	)
	require.NoError(t, err)
	require.False(t, valid)

	// check that empty signatures are not accepted
	valid, err = interpreter.VerifyScriptPubKey(
		script.NewScript(script.P2PK, pubKey),
		"",
		tx1,
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

func TestRPNInterpreter_GenerationAndVerificationRealKeys(t *testing.T) {
	signer := sign.NewECDSASignature()
	interpreter := NewScriptInterpreter(crypto.NewHashedSignature(signer, hash.NewSHA256()))

	pubKey, privKey, err := signer.NewKeyPair()
	require.NoError(t, err)

	// generate the scriptSig to unlock the input
	signature, err := interpreter.GenerateScriptSig(
		script.NewScript(script.P2PK, pubKey),
		privKey,
		tx1,
	)
	require.NoError(t, err)

	// check that the scriptSig generated is correct
	valid, err := interpreter.VerifyScriptPubKey(
		script.NewScript(script.P2PK, pubKey),
		signature,
		tx1,
	)
	require.NoError(t, err)
	assert.True(t, valid)

	// modify the scriptSig and check that is not correct anymore
	modifiedScriptSig := []rune(signature)
	modifiedScriptSig[0] = 'a'
	valid, err = interpreter.VerifyScriptPubKey(
		script.NewScript(script.P2PK, pubKey),
		string(modifiedScriptSig),
		tx1,
	)
	require.NoError(t, err)
	assert.False(t, valid)
}

func TestRPNInterpreter_GenerateScriptSigP2PKMocked(t *testing.T) {
	interpreter := NewScriptInterpreter(crypto.NewHashedSignature(&mockSign.MockSign{}, &mockHash.MockHashing{}))

	// we use a real key pair to generate the signature so the public key can be detected by the interpreter
	// notice that we use the signature mocker, so the signature is predictable and does not depend on key
	pubKey, privKey, err := sign.NewECDSASignature().NewKeyPair()
	require.NoError(t, err)

	signature, err := interpreter.GenerateScriptSig(
		script.NewScript(script.P2PK, pubKey),
		privKey,
		tx1,
	)
	require.NoError(t, err)
	assert.Equal(t, base58.Encode([]byte(fmt.Sprintf("%s-hashed-signed", tx1.AssembleForSigning()))), signature)

	signature, err = interpreter.GenerateScriptSig(
		script.NewScript(script.P2PK, pubKey),
		privKey,
		tx2,
	)
	require.NoError(t, err)
	assert.Equal(t, base58.Encode([]byte(fmt.Sprintf("%s-hashed-signed", tx2.AssembleForSigning()))), signature)

	signature, err = interpreter.GenerateScriptSig(
		script.NewScript(script.P2PK, pubKey),
		privKey,
		tx3,
	)
	require.NoError(t, err)
	assert.Equal(t, base58.Encode([]byte(fmt.Sprintf("%s-hashed-signed", tx3.AssembleForSigning()))), signature)
}

func TestRPNInterpreter_VerifyScriptPubKeyP2PKMocked(t *testing.T) {
	interpreter := NewScriptInterpreter(crypto.NewHashedSignature(&mockSign.MockSign{}, &mockHash.MockHashing{}))

	pubKey, _, err := sign.NewECDSASignature().NewKeyPair()
	require.NoError(t, err)

	valid, err := interpreter.VerifyScriptPubKey(
		script.NewScript(script.P2PK, pubKey),
		base58.Encode([]byte(fmt.Sprintf("%s-hashed-signed", tx1.AssembleForSigning()))),
		tx1,
	)
	require.NoError(t, err)
	assert.True(t, valid)

	valid, err = interpreter.VerifyScriptPubKey(
		script.NewScript(script.P2PK, pubKey),
		base58.Encode([]byte(fmt.Sprintf("%s-hashed-signed", tx2.AssembleForSigning()))),
		tx2,
	)
	require.NoError(t, err)
	assert.True(t, valid)

	valid, err = interpreter.VerifyScriptPubKey(
		script.NewScript(script.P2PK, pubKey),
		base58.Encode([]byte(fmt.Sprintf("%s-hashed-signed", tx3.AssembleForSigning()))),
		tx3,
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
		fmt.Sprintf("%s-hashed-signed", string(tx1.AssembleForSigning())),
		tx1,
	)
	require.NoError(t, err)
	assert.True(t, valid)

	valid, err = interpreter.VerifyScriptPubKey(
		script.NewScript(script.P2PKH, pubKey),
		fmt.Sprintf("%s-hashed-signed", string(tx2.AssembleForSigning())),
		tx2,
	)
	require.NoError(t, err)
	assert.True(t, valid)

	valid, err = interpreter.VerifyScriptPubKey(
		script.NewScript(script.P2PKH, pubKey),
		fmt.Sprintf("%s-hashed-signed", string(tx3.AssembleForSigning())),
		tx3,
	)
	require.NoError(t, err)
	assert.True(t, valid)
}
