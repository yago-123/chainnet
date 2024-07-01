package consensus //nolint:testpackage // don't create separate package for tests
import (
	expl "chainnet/pkg/chain/explorer"
	"chainnet/pkg/kernel"
	"chainnet/pkg/script"
	mockHash "chainnet/tests/mocks/crypto/hash"
	mockSign "chainnet/tests/mocks/crypto/sign"
	mockStorage "chainnet/tests/mocks/storage"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestHValidator_validateInputRemainUnspent(t *testing.T) {

}

func TestHValidator_validateBalance(t *testing.T) {

}

func TestHValidator_validateOwnershipOfInputs(t *testing.T) {

}

func TestHValidator_validateNumberOfCoinbaseTxs(t *testing.T) {
	blockWithoutCoinbase := &kernel.Block{
		Transactions: []*kernel.Transaction{
			kernel.NewTransaction(
				[]kernel.TxInput{kernel.NewInput([]byte("txid"), 0, "scriptSig", "pubKey")},
				[]kernel.TxOutput{kernel.NewOutput(1, script.P2PK, "scriptPubKey")},
			),
		},
	}

	blockWithTwoCoinbase := &kernel.Block{
		Transactions: []*kernel.Transaction{
			kernel.NewCoinbaseTransaction("to"),
			kernel.NewCoinbaseTransaction("to"),
		},
	}

	blockWithOneCoinbase := &kernel.Block{
		Transactions: []*kernel.Transaction{
			kernel.NewCoinbaseTransaction("to"),
		},
	}

	hvalidator := NewHeavyValidator(NewLightValidator(), *expl.NewExplorer(&mockStorage.MockStorage{}), &mockSign.MockSign{}, &mockHash.MockHashing{})

	require.Error(t, hvalidator.validateNumberOfCoinbaseTxs(blockWithoutCoinbase))
	require.Error(t, hvalidator.validateNumberOfCoinbaseTxs(blockWithTwoCoinbase))
	require.NoError(t, hvalidator.validateNumberOfCoinbaseTxs(blockWithOneCoinbase))
}

func TestHValidator_validateNoDoubleSpendingInsideBlock(t *testing.T) {
	blockWithDoubleSpending := &kernel.Block{
		Transactions: []*kernel.Transaction{
			{Vin: []kernel.TxInput{kernel.NewInput([]byte("txid"), 0, "scriptSig", "pubKey")}},
			{Vin: []kernel.TxInput{kernel.NewInput([]byte("txid"), 1, "scriptSig", "pubKey")}},
			{Vin: []kernel.TxInput{kernel.NewInput([]byte("txid"), 0, "scriptSig", "pubKey")}},
		},
	}

	blockWithoutDoubleSpending := &kernel.Block{
		Transactions: []*kernel.Transaction{
			{Vin: []kernel.TxInput{kernel.NewInput([]byte("txid"), 0, "scriptSig", "pubKey")}},
			{Vin: []kernel.TxInput{kernel.NewInput([]byte("txid"), 1, "scriptSig", "pubKey")}},
			{Vin: []kernel.TxInput{kernel.NewInput([]byte("txid2"), 0, "scriptSig", "pubKey")}},
		},
	}

	hvalidator := NewHeavyValidator(NewLightValidator(), *expl.NewExplorer(&mockStorage.MockStorage{}), &mockSign.MockSign{}, &mockHash.MockHashing{})
	require.Error(t, hvalidator.validateNoDoubleSpendingInsideBlock(blockWithDoubleSpending))
	require.NoError(t, hvalidator.validateNoDoubleSpendingInsideBlock(blockWithoutDoubleSpending))
}

func TestHValidator_validateBlockHash(t *testing.T) {

}
