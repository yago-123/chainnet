package validator //nolint:testpackage // don't create separate package for tests

import (
	"github.com/yago-123/chainnet/pkg/common"
	"testing"

	"github.com/yago-123/chainnet/config"

	expl "github.com/yago-123/chainnet/pkg/chain/explorer"
	"github.com/yago-123/chainnet/pkg/consensus"
	"github.com/yago-123/chainnet/pkg/kernel"
	"github.com/yago-123/chainnet/pkg/script"
	"github.com/yago-123/chainnet/pkg/util"
	mockHash "github.com/yago-123/chainnet/tests/mocks/crypto/hash"
	mockSign "github.com/yago-123/chainnet/tests/mocks/crypto/sign"
	mockStorage "github.com/yago-123/chainnet/tests/mocks/storage"

	"github.com/stretchr/testify/require"
)

func TestHValidator_validateOwnershipAndBalanceOfInputs(_ *testing.T) {
	// todo() once we have RPN done
}

func TestHValidator_validateNoCoinbaseAccepted(t *testing.T) {
	hvalidator := NewHeavyValidator(config.NewConfig(), NewLightValidator(&mockHash.FakeHashing{}), expl.NewChainExplorer(&mockStorage.MockStorage{}, &mockHash.FakeHashing{}), &mockSign.MockSign{}, &mockHash.FakeHashing{})

	require.Error(t, hvalidator.ValidateTx(kernel.NewCoinbaseTransaction("to", common.InitialCoinbaseReward, 0)))
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
			kernel.NewCoinbaseTransaction("to", common.InitialCoinbaseReward, 0),
			kernel.NewCoinbaseTransaction("to", common.InitialCoinbaseReward, 0),
		},
	}

	blockWithOneCoinbase := &kernel.Block{
		Transactions: []*kernel.Transaction{
			kernel.NewCoinbaseTransaction("to", common.InitialCoinbaseReward, 0),
		},
	}

	hvalidator := NewHeavyValidator(config.NewConfig(), NewLightValidator(&mockHash.FakeHashing{}), expl.NewChainExplorer(&mockStorage.MockStorage{}, &mockHash.FakeHashing{}), &mockSign.MockSign{}, &mockHash.FakeHashing{})

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
	hvalidator := NewHeavyValidator(config.NewConfig(), NewLightValidator(fakeHashing), expl.NewChainExplorer(&mockStorage.MockStorage{}, &mockHash.FakeHashing{}), &mockSign.MockSign{}, fakeHashing)
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
	hvalidator := NewHeavyValidator(config.NewConfig(), NewLightValidator(fakeHashing), expl.NewChainExplorer(&mockStorage.MockStorage{}, &mockHash.FakeHashing{}), &mockSign.MockSign{}, fakeHashing)

	// check that the block hash corresponds to the target
	require.NoError(t, hvalidator.validateBlockHash(block))

	// check that the block hash does not correspond to the target
	block.Header.SetNonce(2)
	require.Error(t, hvalidator.validateBlockHash(block))
}

func TestHValidator_validateHeaderPreviousBlock(t *testing.T) {
	mockStore := &mockStorage.MockStorage{}

	mockHeader := &kernel.BlockHeader{MerkleRoot: []byte("merkle root")}
	mockStore.
		On("GetLastHeader").
		Return(mockHeader, nil)
	fakeHashing := &mockHash.FakeHashing{}
	hvalidator := NewHeavyValidator(config.NewConfig(), NewLightValidator(fakeHashing), expl.NewChainExplorer(mockStore, fakeHashing), &mockSign.MockSign{}, fakeHashing)

	// check that the previous block hash of the block matches the latest block
	require.NoError(t, hvalidator.validateHeaderPreviousBlock(&kernel.BlockHeader{PrevBlockHash: append(mockHeader.Assemble(), []byte("-hashed")...), Height: 1}))

	// check that the previous block hash of the block does not match the latest block
	require.Error(t, hvalidator.validateHeaderPreviousBlock(&kernel.BlockHeader{PrevBlockHash: []byte("block-2"), Height: 1}))

	// check that genesis block does not validate the previous block hash
	require.NoError(t, hvalidator.validateHeaderPreviousBlock(&kernel.BlockHeader{PrevBlockHash: []byte{}, Height: 0}))

	// check that genesis block requires empty previous block hash
	require.Error(t, hvalidator.validateHeaderPreviousBlock(&kernel.BlockHeader{PrevBlockHash: []byte("block-1"), Height: 0}))
}

func TestHValidator_validateGenesisHeader(t *testing.T) {
	mockStore := &mockStorage.MockStorage{}

	mockHeader := &kernel.BlockHeader{MerkleRoot: []byte("merkle root")}
	mockStore.
		On("GetLastHeader").
		Return(mockHeader, nil)
	fakeHashing := &mockHash.FakeHashing{}
	hvalidator := NewHeavyValidator(config.NewConfig(), NewLightValidator(fakeHashing), expl.NewChainExplorer(mockStore, fakeHashing), &mockSign.MockSign{}, fakeHashing)

	// check that can be a single genesis block
	require.Error(t, hvalidator.validateGenesisHeader(&kernel.BlockHeader{Height: 0, PrevBlockHash: []byte{}}))
}

func TestHValidator_validateBlockHeight(t *testing.T) {
	mockStore := &mockStorage.MockStorage{}
	mockStore.
		On("GetLastHeader").
		Return(&kernel.BlockHeader{Height: 10}, nil)

	fakeHashing := &mockHash.FakeHashing{}
	hvalidator := NewHeavyValidator(config.NewConfig(), NewLightValidator(fakeHashing), expl.NewChainExplorer(mockStore, &mockHash.FakeHashing{}), &mockSign.MockSign{}, fakeHashing)

	// check that the block height matches the current chain height
	require.NoError(t, hvalidator.validateHeaderHeight(&kernel.BlockHeader{Height: 11}))

	// check that the block height does not match the current chain height
	require.Error(t, hvalidator.validateHeaderHeight(&kernel.BlockHeader{Height: 10}))
	require.Error(t, hvalidator.validateHeaderHeight(&kernel.BlockHeader{Height: 12}))
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

	hvalidator := NewHeavyValidator(config.NewConfig(), NewLightValidator(&mockHash.FakeHashing{}), expl.NewChainExplorer(&mockStorage.MockStorage{}, &mockHash.FakeHashing{}), &mockSign.MockSign{}, &mockHash.FakeHashing{})

	// verify correct merkle root does not generate error
	require.NoError(t, hvalidator.validateMerkleTree(block))

	// verify incorrect merkle root generates error
	block.Transactions[0].Vin[0].Txid = []byte("invalid")
	require.Error(t, hvalidator.validateMerkleTree(block))
}
