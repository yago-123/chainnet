package validator //nolint:testpackage // don't create separate package for tests

import (
	"github.com/yago-123/chainnet/pkg/kernel"
	"github.com/yago-123/chainnet/pkg/script"
	"github.com/yago-123/chainnet/tests/mocks/crypto/hash"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLValidator_ValidateTxLight(_ *testing.T) {
	// todo() once we have RPN done
}

func TestLValidator_ValidateHeader(_ *testing.T) {

}

func TestLValidator_validateTxID(t *testing.T) {
	hasher := &hash.FakeHashing{}
	tx := kernel.NewTransaction(
		[]kernel.TxInput{kernel.NewInput([]byte("tx-id-1"), 1, "scriptsig", "pubkey")},
		[]kernel.TxOutput{kernel.NewOutput(1, script.P2PK, "pubkey2")},
	)

	// generate tx hash
	txHash, err := hasher.Hash(tx.Assemble())
	require.NoError(t, err)

	lv := LValidator{
		hasher: hasher,
	}
	// verify that tx hash matches
	tx.SetID(txHash)
	require.NoError(t, lv.validateTxID(tx))

	// modify the hash
	txHash[0] = 0x0
	tx.SetID(txHash)
	require.Error(t, lv.validateTxID(tx))
}
