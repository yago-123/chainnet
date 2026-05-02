//go:build e2e

package e2e

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net"
	"path/filepath"
	"testing"
	"time"

	"github.com/btcsuite/btcutil/base58"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"github.com/yago-123/chainnet/config"
	blockchain "github.com/yago-123/chainnet/pkg/chain"
	"github.com/yago-123/chainnet/pkg/chain/explorer"
	"github.com/yago-123/chainnet/pkg/consensus/validator"
	"github.com/yago-123/chainnet/pkg/crypto/hash"
	"github.com/yago-123/chainnet/pkg/crypto/sign"
	"github.com/yago-123/chainnet/pkg/encoding"
	"github.com/yago-123/chainnet/pkg/kernel"
	"github.com/yago-123/chainnet/pkg/mempool"
	"github.com/yago-123/chainnet/pkg/miner"
	"github.com/yago-123/chainnet/pkg/network"
	"github.com/yago-123/chainnet/pkg/observer"
	"github.com/yago-123/chainnet/pkg/script"
	"github.com/yago-123/chainnet/pkg/script/interpreter"
	sdkv1beta "github.com/yago-123/chainnet/pkg/sdk/v1beta"
	"github.com/yago-123/chainnet/pkg/storage"
	"github.com/yago-123/chainnet/pkg/util"
	"github.com/yago-123/chainnet/pkg/utxoset"
)

const (
	e2eBlockBucket  = "sdk-e2e-blocks"
	e2eHeaderBucket = "sdk-e2e-headers"
)

func TestSDK_SubmitTransactionsMineBlockAndReadBack(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cfg := newSDKTestConfig(t)
	hasher := hash.NewHasher(sha256.New())
	signer := testSignature{}
	pubKey, privKey, err := signer.NewKeyPair()
	require.NoError(t, err)
	cfg.Miner.PubKey = base58.Encode(pubKey)

	store, err := storage.NewBoltDB(
		filepath.Join(t.TempDir(), "chainnet.db"),
		e2eBlockBucket,
		e2eHeaderBucket,
		encoding.NewGobEncoder(),
	)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, store.Close()) })

	chainSubject := observer.NewChainSubject()
	netSubject := observer.NewNetSubject()
	memPool := mempool.NewMemPool(100)
	utxoSet := utxoset.NewUTXOSet(cfg)
	chainExplorer := explorer.NewChainExplorer(store, hasher)
	heavyValidator := validator.NewHeavyValidator(
		cfg,
		validator.NewLightValidator(hasher),
		chainExplorer,
		signer,
		hasher,
	)

	chain, err := blockchain.NewBlockchain(
		cfg,
		store,
		memPool,
		utxoSet,
		hasher,
		heavyValidator,
		chainSubject,
		encoding.NewGobEncoder(),
	)
	require.NoError(t, err)

	chainSubject.Register(store)
	chainSubject.Register(memPool)
	chainSubject.Register(utxoSet)
	netSubject.Register(chain)

	blockMiner, err := miner.NewMiner(cfg, chain, hash.SHA256, chainExplorer)
	require.NoError(t, err)

	_, err = blockMiner.MineBlock()
	require.NoError(t, err)
	fundingBlock1, err := blockMiner.MineBlock()
	require.NoError(t, err)
	fundingBlock2, err := blockMiner.MineBlock()
	require.NoError(t, err)

	router := network.NewHTTPRouter(cfg, chainExplorer, netSubject)
	require.NoError(t, router.Start())
	t.Cleanup(func() { require.NoError(t, router.Stop()) })

	client, err := sdkv1beta.NewClient(fmt.Sprintf("127.0.0.1:%d", cfg.P2P.RouterPort), nil)
	require.NoError(t, err)
	require.Eventually(t, func() bool {
		_, err := client.GetLatestHeader(ctx)
		return err == nil
	}, 5*time.Second, 50*time.Millisecond)

	tx1 := newSignedSpendTx(t, signer, hasher, pubKey, privKey, fundingBlock1.Transactions[0], 0, []byte("sdk-e2e-recipient-1"), 10, 1)
	tx2 := newSignedSpendTx(t, signer, hasher, pubKey, privKey, fundingBlock2.Transactions[0], 0, []byte("sdk-e2e-recipient-2"), 20, 1)

	require.NoError(t, client.SendTransaction(ctx, *tx1))
	require.NoError(t, client.SendTransaction(ctx, *tx2))

	minedBlock, err := blockMiner.MineBlock()
	require.NoError(t, err)
	require.True(t, containsTxID(minedBlock, tx1.ID), "mined block does not contain tx1")
	require.True(t, containsTxID(minedBlock, tx2.ID), "mined block does not contain tx2")

	latestBlock, err := client.GetLatestBlock(ctx)
	require.NoError(t, err)
	require.Equal(t, minedBlock.Hash, latestBlock.Hash)
	require.True(t, containsTxID(latestBlock, tx1.ID), "latest block does not contain tx1")
	require.True(t, containsTxID(latestBlock, tx2.ID), "latest block does not contain tx2")

	blockByHash, err := client.GetBlockByHash(ctx, minedBlock.Hash)
	require.NoError(t, err)
	require.Equal(t, minedBlock.Hash, blockByHash.Hash)

	gotTx1, err := client.GetTransactionByID(ctx, tx1.ID)
	require.NoError(t, err)
	require.Equal(t, tx1.ID, gotTx1.ID)

	gotTx2, err := client.GetTransactionByID(ctx, tx2.ID)
	require.NoError(t, err)
	require.Equal(t, tx2.ID, gotTx2.ID)

	latestHeader, err := client.GetLatestHeader(ctx)
	require.NoError(t, err)
	require.Equal(t, minedBlock.Header.Height, latestHeader.Height)

	active, err := client.AddressIsActive(ctx, []byte("sdk-e2e-recipient-1"))
	require.NoError(t, err)
	require.True(t, active)

	recipientTxs, err := client.GetAddressTransactions(ctx, []byte("sdk-e2e-recipient-1"))
	require.NoError(t, err)
	require.NotEmpty(t, recipientTxs)
}

func newSDKTestConfig(t *testing.T) *config.Config {
	t.Helper()

	port := freeTCPPort(t)
	logger := logrus.New()
	logger.SetOutput(io.Discard)

	cfg := config.NewConfig()
	cfg.Logger = logger
	cfg.P2P.Enabled = false
	cfg.P2P.RouterPort = uint(port)
	cfg.P2P.ReadTimeout = 5 * time.Second
	cfg.P2P.WriteTimeout = 5 * time.Second
	cfg.P2P.ConnTimeout = 5 * time.Second
	cfg.Wallet.ServerAddress = "127.0.0.1"
	cfg.Wallet.ServerPort = uint(port)

	return cfg
}

func freeTCPPort(t *testing.T) int {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer listener.Close()

	return listener.Addr().(*net.TCPAddr).Port
}

func newSignedSpendTx(
	t *testing.T,
	signer sign.Signature,
	hasher hash.Hashing,
	pubKey []byte,
	privKey []byte,
	sourceTx *kernel.Transaction,
	sourceOutIdx uint,
	recipient []byte,
	amount uint,
	fee uint,
) *kernel.Transaction {
	t.Helper()

	sourceOutput := sourceTx.Vout[sourceOutIdx]
	change := sourceOutput.Amount - amount - fee
	tx := kernel.NewTransaction(
		[]kernel.TxInput{
			kernel.NewInput(sourceTx.ID, sourceOutIdx, "", string(pubKey)),
		},
		[]kernel.TxOutput{
			kernel.NewOutput(amount, script.P2PK, string(recipient)),
			kernel.NewOutput(change, script.P2PK, string(pubKey)),
		},
	)

	scriptSig, err := interpreter.NewScriptInterpreter(signer).GenerateScriptSig(sourceOutput.ScriptPubKey, pubKey, privKey, tx)
	require.NoError(t, err)
	tx.Vin[0].ScriptSig = scriptSig

	txID, err := util.CalculateTxHash(tx, hasher)
	require.NoError(t, err)
	tx.SetID(txID)

	return tx
}

type testSignature struct{}

func (testSignature) NewKeyPair() ([]byte, []byte, error) {
	return []byte("sdk-e2e-owner-public-key"), []byte("sdk-e2e-owner-private-key"), nil
}

func (testSignature) Sign(payload []byte, _ []byte) ([]byte, error) {
	return []byte("sdk-e2e-signature-" + hex.EncodeToString(payload)), nil
}

func (testSignature) Verify(signature []byte, payload []byte, _ []byte) (bool, error) {
	expected, err := testSignature{}.Sign(payload, nil)
	if err != nil {
		return false, err
	}

	return bytes.Equal(signature, expected), nil
}

func containsTxID(block *kernel.Block, txID []byte) bool {
	for _, tx := range block.Transactions {
		if bytes.Equal(tx.ID, txID) {
			return true
		}
	}

	return false
}
