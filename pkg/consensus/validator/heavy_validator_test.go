package validator //nolint:testpackage // don't create separate package for tests

import (
	expl "chainnet/pkg/chain/explorer"
	"chainnet/pkg/consensus"
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

	hvalidator := NewHeavyValidator(NewLightValidator(&mockHash.FakeHashing{}), expl.NewExplorer(&mockStorage.MockStorage{}, &mockHash.MockHashing{}), &mockSign.MockSign{}, &mockHash.FakeHashing{})

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

	fakeHashing := &mockHash.FakeHashing{}
	hvalidator := NewHeavyValidator(NewLightValidator(fakeHashing), expl.NewExplorer(&mockStorage.MockStorage{}, fakeHashing), &mockSign.MockSign{}, fakeHashing)
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

	blockHash, err := util.CalculateBlockHash(blockHeader, &mockHash.FakeHashing{})
	require.NoError(t, err)
	block := &kernel.Block{
		Header:       blockHeader,
		Transactions: []*kernel.Transaction{},
		Hash:         blockHash,
	}

	fakeHashing := &mockHash.FakeHashing{}
	hvalidator := NewHeavyValidator(NewLightValidator(fakeHashing), expl.NewExplorer(&mockStorage.MockStorage{}, fakeHashing), &mockSign.MockSign{}, fakeHashing)

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
	fakeHashing := &mockHash.FakeHashing{}
	hvalidator := NewHeavyValidator(NewLightValidator(fakeHashing), expl.NewExplorer(mockStore, fakeHashing), &mockSign.MockSign{}, fakeHashing)

	// check that the previous block hash of the block matches the latest block
	require.NoError(t, hvalidator.validatePreviousBlockMatchCurrentLatest(&kernel.Block{Header: &kernel.BlockHeader{PrevBlockHash: []byte("block-1"), Height: 1}}))

	// check that the previous block hash of the block does not match the latest block
	require.Error(t, hvalidator.validatePreviousBlockMatchCurrentLatest(&kernel.Block{Header: &kernel.BlockHeader{PrevBlockHash: []byte("block-2"), Height: 1}}))

	// check that genesis block does not validate the previous block hash
	require.NoError(t, hvalidator.validatePreviousBlockMatchCurrentLatest(&kernel.Block{Header: &kernel.BlockHeader{PrevBlockHash: []byte{}, Height: 0}}))

	// check that genesis block requires empty previous block hash
	require.Error(t, hvalidator.validatePreviousBlockMatchCurrentLatest(&kernel.Block{Header: &kernel.BlockHeader{PrevBlockHash: []byte("block-1"), Height: 0}}))
}

func TestHValidator_validateBlockHeight(t *testing.T) {
	mockStore := &mockStorage.MockStorage{}
	mockStore.
		On("GetLastBlock").
		Return(&kernel.Block{Header: &kernel.BlockHeader{Height: 10}}, nil)

	fakeHashing := &mockHash.FakeHashing{}
	hvalidator := NewHeavyValidator(NewLightValidator(fakeHashing), expl.NewExplorer(mockStore, fakeHashing), &mockSign.MockSign{}, fakeHashing)

	// check that the block height matches the current chain height
	require.NoError(t, hvalidator.validateBlockHeight(&kernel.Block{Header: &kernel.BlockHeader{Height: 11}}))

	// check that the block height does not match the current chain height
	require.Error(t, hvalidator.validateBlockHeight(&kernel.Block{Header: &kernel.BlockHeader{Height: 10}}))
	require.Error(t, hvalidator.validateBlockHeight(&kernel.Block{Header: &kernel.BlockHeader{Height: 12}}))
}

func TestHValidator_validateMerkleTree(t *testing.T) {
	txs := []*kernel.Transaction{
		kernel.NewTransaction(
			[]kernel.TxInput{kernel.NewInput([]byte("txid"), 0, "scriptSig", "pubKey")},
			[]kernel.TxOutput{kernel.NewOutput(1, script.P2PK, "scriptPubKey")},
		),
		kernel.NewTransaction(
			[]kernel.TxInput{kernel.NewInput([]byte("txid2"), 1, "scriptSig2", "pubKey2")},
			[]kernel.TxOutput{
				kernel.NewOutput(2, script.P2PK, "scriptPubKey2"),
				kernel.NewOutput(2, script.P2PK, "scriptPubKey3"),
			},
		),
		kernel.NewTransaction(
			[]kernel.TxInput{
				kernel.NewInput([]byte("txid3"), 2, "scriptSig3", "pubKey3"),
				kernel.NewInput([]byte("txid4"), 3, "scriptSig4", "pubKey4"),
			},
			[]kernel.TxOutput{
				kernel.NewOutput(3, script.P2PK, "scriptPubKey3"),
			},
		),
	}

	for _, tx := range txs {
		txHash, err := util.CalculateTxHash(tx, &mockHash.FakeHashing{})
		require.NoError(t, err)
		tx.SetID(txHash)
	}

	mt, err := consensus.NewMerkleTreeFromTxs(txs, &mockHash.FakeHashing{})
	require.NoError(t, err)

	block := &kernel.Block{
		Header: &kernel.BlockHeader{
			MerkleRoot: mt.RootHash(),
		},
		Transactions: txs,
	}

	hvalidator := NewHeavyValidator(NewLightValidator(&mockHash.FakeHashing{}), expl.NewExplorer(&mockStorage.MockStorage{}, &mockHash.FakeHashing{}), &mockSign.MockSign{}, &mockHash.FakeHashing{})

	// verify correct merkle root does not generate error
	require.NoError(t, hvalidator.validateMerkleTree(block))

	// verify incorrect merkle root generates error
	block.Transactions[0].Vin[0].Txid = []byte("invalid")
	require.Error(t, hvalidator.validateMerkleTree(block))
}
