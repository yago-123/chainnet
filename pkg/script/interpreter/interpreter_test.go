package interpreter //nolint:testpackage // don't create separate package for tests
import (
	"chainnet/pkg/crypto"
	"chainnet/pkg/crypto/sign"
	"chainnet/pkg/kernel"
	"chainnet/pkg/script"
	mockHash "chainnet/tests/mocks/crypto/hash"
	mockSign "chainnet/tests/mocks/crypto/sign"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

// tx1 contains a single input and a single output
var tx1 = kernel.NewTransaction(
	[]kernel.TxInput{
		kernel.NewInput([]byte("transaction-1"), 1, "", "pubKey-1"),
	},
	[]kernel.TxOutput{
		kernel.NewOutput(50, script.P2PK, "pubKey-1"),
	},
)

// tx2 contains multiple inputs with same public key
var tx2 = kernel.NewTransaction(
	[]kernel.TxInput{
		kernel.NewInput([]byte("transaction-1"), 1, "", "pubKey-1"),
		kernel.NewInput([]byte("transaction-2"), 1, "", "pubKey-1"),
	},
	[]kernel.TxOutput{
		kernel.NewOutput(50, script.P2PK, "pubKey-1"),
	},
)

// tx3 contains multiple inputs with different public keys
var tx3 = kernel.NewTransaction(
	[]kernel.TxInput{
		kernel.NewInput([]byte("transaction-1"), 1, "", "pubKey-1"),
		kernel.NewInput([]byte("transaction-2"), 1, "", "pubKey-2"),
	},
	[]kernel.TxOutput{
		kernel.NewOutput(50, script.P2PK, "pubKey-1"),
	},
)

func TestRPNInterpreter_GenerateScriptSigWithErrors(t *testing.T) {

}

func TestRPNInterpreter_VerifyScriptPubKeyWithErrors(t *testing.T) {

}

func TestRPNInterpreter_GenerateScriptSigP2PKMocked(t *testing.T) {
	interpreter := NewScriptInterpreter(crypto.NewHashedSignature(&mockSign.MockSign{}, &mockHash.MockHashing{}))

	// we use a real key pair to generate the signature so the public key can be detected by the interpreter
	// notice that we use the signature mocker, so the signature is predictable and does not depend on key
	pubKey, privKey, _ := sign.NewECDSASignature().NewKeyPair()

	signature, err := interpreter.GenerateScriptSig(
		script.NewScript(script.P2PK, pubKey),
		privKey,
		tx1,
	)
	require.NoError(t, err)
	assert.Equal(t, fmt.Sprintf("%s-hashed-signed", string(tx1.AssembleForSigning())), signature)

	signature, err = interpreter.GenerateScriptSig(
		script.NewScript(script.P2PK, pubKey),
		privKey,
		tx2,
	)
	require.NoError(t, err)
	assert.Equal(t, fmt.Sprintf("%s-hashed-signed", string(tx2.AssembleForSigning())), signature)

	signature, err = interpreter.GenerateScriptSig(
		script.NewScript(script.P2PK, pubKey),
		privKey,
		tx3,
	)
	require.NoError(t, err)
	assert.Equal(t, fmt.Sprintf("%s-hashed-signed", string(tx3.AssembleForSigning())), signature)
}

func TestRPNInterpreter_VerifyScriptPubKeyP2PKMocked(t *testing.T) {
	interpreter := NewScriptInterpreter(crypto.NewHashedSignature(&mockSign.MockSign{}, &mockHash.MockHashing{}))

	// we use a real key pair to generate the signature so the public key can be detected by the interpreter
	// notice that we use the signature mocker, so the signature is predictable and does not depend on key
	pubKey, _, _ := sign.NewECDSASignature().NewKeyPair()

	valid, err := interpreter.VerifyScriptPubKey(
		script.NewScript(script.P2PK, pubKey),
		fmt.Sprintf("%s-hashed-signed", string(tx1.AssembleForSigning())),
		tx1,
	)
	require.NoError(t, err)
	assert.True(t, valid)

	valid, err = interpreter.VerifyScriptPubKey(
		script.NewScript(script.P2PK, pubKey),
		fmt.Sprintf("%s-hashed-signed", string(tx2.AssembleForSigning())),
		tx2,
	)
	require.NoError(t, err)
	assert.True(t, valid)

	valid, err = interpreter.VerifyScriptPubKey(
		script.NewScript(script.P2PK, pubKey),
		fmt.Sprintf("%s-hashed-signed", string(tx3.AssembleForSigning())),
		tx3,
	)
	require.NoError(t, err)
	assert.True(t, valid)

}
