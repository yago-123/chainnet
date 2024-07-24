package validator //nolint:testpackage // don't create separate package for tests

import (
	expl "chainnet/pkg/chain/explorer"
	"chainnet/pkg/consensus/miner"
	"chainnet/pkg/consensus/util"
	"chainnet/pkg/kernel"
	"chainnet/pkg/script"
	mockHash "chainnet/tests/mocks/crypto/hash"
	mockSign "chainnet/tests/mocks/crypto/sign"
	mockStorage "chainnet/tests/mocks/storage"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHValidator_validateOwnershipAndBalanceOfInputs(_ *testing.T) {
	// todo() once we have RPN done
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
			kernel.NewCoinbaseTransaction("to", miner.InitialCoinbaseReward, 0),
			kernel.NewCoinbaseTransaction("to", miner.InitialCoinbaseReward, 0),
		},
	}

	blockWithOneCoinbase := &kernel.Block{
		Transactions: []*kernel.Transaction{
			kernel.NewCoinbaseTransaction("to", miner.InitialCoinbaseReward, 0),
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
	blockHeader := &kernel.BlockHeader{
		Version:       []byte("version"),
		PrevBlockHash: []byte("prevBlockHash"),
		MerkleRoot:    []byte("merkleRoot"),
		Height:        1,
		Timestamp:     2,
		Target:        3,
		Nonce:         4,
	}

	blockHash, err := util.CalculateBlockHash(blockHeader, &mockHash.MockHashing{})
	require.NoError(t, err)
	block := &kernel.Block{
		Header:       blockHeader,
		Transactions: []*kernel.Transaction{},
		Hash:         blockHash,
	}

	hvalidator := NewHeavyValidator(NewLightValidator(), *expl.NewExplorer(&mockStorage.MockStorage{}), &mockSign.MockSign{}, &mockHash.MockHashing{})

	// check that the block hash corresponds to the target
	require.NoError(t, hvalidator.validateBlockHash(block))

	// check that the block hash does not correspond to the target
	block.Header.SetNonce(2)
	require.Error(t, hvalidator.validateBlockHash(block))
}

func TestHValidator_validatePreviousBlockMatchCurrentLatest(t *testing.T) {
	mockStore := &mockStorage.MockStorage{}
	mockStore.
		On("GetLastBlock").
		Return(&kernel.Block{Hash: []byte("block-1")}, nil)
	hvalidator := NewHeavyValidator(NewLightValidator(), *expl.NewExplorer(mockStore), &mockSign.MockSign{}, &mockHash.MockHashing{})

	// check that the previous block hash of the block matches the latest block
	require.NoError(t, hvalidator.validatePreviousBlockMatchCurrentLatest(&kernel.Block{Header: &kernel.BlockHeader{PrevBlockHash: []byte("block-1")}}))

	// check that the previous block hash of the block does not match the latest block
	require.Error(t, hvalidator.validatePreviousBlockMatchCurrentLatest(&kernel.Block{Header: &kernel.BlockHeader{PrevBlockHash: []byte("block-2")}}))
}

func TestHValidator_validateBlockHeight(t *testing.T) {
	mockStore := &mockStorage.MockStorage{}
	mockStore.
		On("GetLastBlock").
		Return(&kernel.Block{Header: &kernel.BlockHeader{Height: 10}}, nil)

	hvalidator := NewHeavyValidator(NewLightValidator(), *expl.NewExplorer(mockStore), &mockSign.MockSign{}, &mockHash.MockHashing{})

	// check that the block height matches the current chain height
	require.NoError(t, hvalidator.validateBlockHeight(&kernel.Block{Header: &kernel.BlockHeader{Height: 11}}))

	// check that the block height does not match the current chain height
	require.Error(t, hvalidator.validateBlockHeight(&kernel.Block{Header: &kernel.BlockHeader{Height: 10}}))
	require.Error(t, hvalidator.validateBlockHeight(&kernel.Block{Header: &kernel.BlockHeader{Height: 12}}))
}

func TestHValidator_validateMerkleTree(_ *testing.T) {
	// todo(): add tests regarding Merkle tree
}
