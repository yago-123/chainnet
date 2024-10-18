package miner //nolint:testpackage // don't create separate package for tests

import (
	"chainnet/config"
	blockchain "chainnet/pkg/chain"
	"chainnet/pkg/chain/explorer"
	"chainnet/pkg/consensus/util"
	"chainnet/pkg/crypto/hash"
	"chainnet/pkg/encoding"
	"chainnet/pkg/kernel"
	"chainnet/pkg/mempool"
	"chainnet/pkg/observer"
	"chainnet/pkg/script"
	"chainnet/pkg/storage"
	"chainnet/tests/mocks/consensus"
	mockStorage "chainnet/tests/mocks/storage"
	"context"
	"testing"

	"github.com/sirupsen/logrus"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var txFeePair1 = mempool.TxFeePair{ //nolint: gochecknoglobals // no need to lint this global variable
	Transaction: &kernel.Transaction{
		ID:   []byte("id1"),
		Vin:  []kernel.TxInput{kernel.NewInput([]byte("txid"), 0, "scriptSig", "pubKey")},
		Vout: []kernel.TxOutput{kernel.NewOutput(1, script.P2PK, "scriptPubKey")},
	},
	Fee: 10,
}

var txFeePair2 = mempool.TxFeePair{ //nolint: gochecknoglobals // no need to lint this global variable
	Transaction: &kernel.Transaction{
		ID:   []byte("id2"),
		Vin:  []kernel.TxInput{kernel.NewInput([]byte("txid2"), 1, "scriptSig", "pubKey")},
		Vout: []kernel.TxOutput{kernel.NewOutput(2, script.P2PK, "scriptPubKey")},
	},
	Fee: 2,
}

var txFeePair3 = mempool.TxFeePair{ //nolint: gochecknoglobals // no need to lint this global variable
	Transaction: &kernel.Transaction{
		ID:   []byte("id3"),
		Vin:  []kernel.TxInput{kernel.NewInput([]byte("txid3"), 2, "scriptSig", "pubKey")},
		Vout: []kernel.TxOutput{kernel.NewOutput(3, script.P2PK, "scriptPubKey")},
	},
}

var txFeePair4 = mempool.TxFeePair{ //nolint: gochecknoglobals // no need to lint this global variable
	Transaction: &kernel.Transaction{
		ID:   []byte("id4"),
		Vin:  []kernel.TxInput{kernel.NewInput([]byte("txid4"), 3, "scriptSig", "pubKey")},
		Vout: []kernel.TxOutput{kernel.NewOutput(4, script.P2PK, "scriptPubKey")},
	},
	Fee: 1,
}

var txFeePair5 = mempool.TxFeePair{ //nolint: gochecknoglobals // no need to lint this global variable
	Transaction: &kernel.Transaction{
		ID:   []byte("id5"),
		Vin:  []kernel.TxInput{kernel.NewInput([]byte("txid5"), 4, "scriptSig", "pubKey")},
		Vout: []kernel.TxOutput{kernel.NewOutput(5, script.P2PK, "scriptPubKey")},
	},
	Fee: 9,
}

var txFeePair6 = mempool.TxFeePair{ //nolint: gochecknoglobals // no need to lint this global variable
	Transaction: &kernel.Transaction{
		ID:   []byte("id6"),
		Vin:  []kernel.TxInput{kernel.NewInput([]byte("txid6"), 5, "scriptSig", "pubKey")},
		Vout: []kernel.TxOutput{kernel.NewOutput(6, script.P2PK, "scriptPubKey")},
	},
	Fee: 6,
}

var txs = []mempool.TxFeePair{txFeePair1, txFeePair2, txFeePair3, txFeePair4, txFeePair5, txFeePair6} //nolint: gochecknoglobals // no need to lint this global variable

func TestMiner_MineBlock(t *testing.T) {
	mempool := mempool.NewMemPool()
	for _, v := range txs {
		txID, err := util.CalculateTxHash(v.Transaction, hash.NewSHA256())
		require.NoError(t, err)

		v.Transaction.SetID(txID)
		mempool.AppendTransaction(v.Transaction, v.Fee)
	}

	store := &mockStorage.MockStorage{}
	store.
		On("GetLastHeader").
		Return(&kernel.BlockHeader{}, storage.ErrNotFound)
	store.
		On("GetLastBlockHash").
		Return([]byte{}, storage.ErrNotFound)

	cfg := config.NewConfig()
	chain, err := blockchain.NewBlockchain(cfg, store, mempool, hash.NewSHA256(), consensus.NewMockHeavyValidator(), observer.NewChainSubject(), encoding.NewGobEncoder())
	require.NoError(t, err)

	miner := Miner{
		hasherType:  hash.SHA256,
		minerPubKey: []byte("minerPubKey"),
		chain:       chain,
		cfg:         cfg,
	}

	// simple block mining with hash difficulty 16
	block, err := miner.MineBlock()
	require.NoError(t, err)
	assert.Len(t, block.Transactions, len(txs)+1)
	assert.True(t, block.Transactions[0].IsCoinbase())
	assert.Positive(t, block.Header.Nonce)
	assert.Equal(t, script.NewScript(script.P2PK, []byte("minerPubKey")), block.Transactions[0].Vout[0].ScriptPubKey)
	assert.Equal(t, byte(0), block.Hash[0]&0x80)

	// cancel block in the middle of mining aborting the process
	miner.ctx, miner.cancel = context.WithCancel(context.Background())
	miner.isMining = true
	miner.cancel()
	_, err = miner.MineBlock()
	require.Error(t, err)
}

func TestMiner_MineBlockWithTargetChange(t *testing.T) {

}

func TestMiner_createCoinbaseTransaction(t *testing.T) {
	store := &mockStorage.MockStorage{}
	store.
		On("GetLastHeader").
		Return(&kernel.BlockHeader{}, storage.ErrNotFound)
	store.
		On("GetLastBlockHash").
		Return([]byte{}, storage.ErrNotFound)

	cfg := config.NewConfig()
	chain, err := blockchain.NewBlockchain(&config.Config{Logger: logrus.New()}, store, mempool.NewMemPool(), hash.NewSHA256(), consensus.NewMockHeavyValidator(), observer.NewChainSubject(), encoding.NewGobEncoder())
	require.NoError(t, err)

	cfg.PubKey = "12D3KooWACTzxPJTeyuFKDQQnzZs3WrynJ6L67BZGPCKAgZrNzZe"
	miner, err := NewMiner(cfg, chain, hash.SHA256, explorer.NewExplorer(store, hash.GetHasher(hash.SHA256)))
	require.NoError(t, err)

	coinbase, err := miner.createCoinbaseTransaction(0, 0)
	require.NoError(t, err)
	assert.Len(t, coinbase.Vout, 1)
	assert.NotEmpty(t, coinbase.ID)
	assert.Equal(t, uint(InitialCoinbaseReward), coinbase.Vout[0].Amount)

	coinbase, err = miner.createCoinbaseTransaction(1, 1)
	require.NoError(t, err)
	assert.Len(t, coinbase.Vout, 2)
	assert.NotEmpty(t, coinbase.ID)
	assert.Equal(t, uint(InitialCoinbaseReward), coinbase.Vout[0].Amount)
	assert.Equal(t, uint(1), coinbase.Vout[1].Amount)

	coinbase, err = miner.createCoinbaseTransaction(0, HalvingInterval)
	require.NoError(t, err)
	assert.Len(t, coinbase.Vout, 1)
	assert.NotEmpty(t, coinbase.ID)
	assert.Equal(t, uint(InitialCoinbaseReward/2), coinbase.Vout[0].Amount)

	coinbase, err = miner.createCoinbaseTransaction(0, HalvingInterval*2)
	require.NoError(t, err)
	assert.Len(t, coinbase.Vout, 1)
	assert.NotEmpty(t, coinbase.ID)
	assert.Equal(t, uint(InitialCoinbaseReward/4), coinbase.Vout[0].Amount)

	coinbase, err = miner.createCoinbaseTransaction(0, HalvingInterval*64)
	require.NoError(t, err)
	assert.Len(t, coinbase.Vout, 1)
	assert.NotEmpty(t, coinbase.ID)
	assert.Equal(t, uint(0), coinbase.Vout[0].Amount)
}
