package interpreter //nolint:testpackage // don't create separate package for tests
import (
	"chainnet/pkg/kernel"
	"chainnet/pkg/script"
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

func TestRPNInterpreter_GenerateScriptSigP2PK(t *testing.T) {

}

func TestRPNInterpreter_VerifyScriptPubKeyP2PK(t *testing.T) {

}
