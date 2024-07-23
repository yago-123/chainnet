package miner //nolint:testpackage // don't create separate package for tests

import (
	"chainnet/pkg/crypto/hash"
	"chainnet/pkg/kernel"
	"chainnet/pkg/script"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var txFeePair1 = txFeePair{ //nolint: gochecknoglobals // no need to lint this global variable
	Transaction: &kernel.Transaction{ID: []byte("id1")},
	Fee:         10,
}

var txFeePair2 = txFeePair{ //nolint: gochecknoglobals // no need to lint this global variable
	Transaction: &kernel.Transaction{ID: []byte("id2")},
	Fee:         2,
}

var txFeePair3 = txFeePair{ //nolint: gochecknoglobals // no need to lint this global variable
	Transaction: &kernel.Transaction{ID: []byte("id3")},
}

var txFeePair4 = txFeePair{ //nolint: gochecknoglobals // no need to lint this global variable
	Transaction: &kernel.Transaction{ID: []byte("id4")},
	Fee:         1,
}

var txFeePair5 = txFeePair{ //nolint: gochecknoglobals // no need to lint this global variable
	Transaction: &kernel.Transaction{ID: []byte("id5")},
	Fee:         9,
}

var txFeePair6 = txFeePair{ //nolint: gochecknoglobals // no need to lint this global variable
	Transaction: &kernel.Transaction{ID: []byte("id6")},
	Fee:         6,
}

var txs = []txFeePair{txFeePair1, txFeePair2, txFeePair3, txFeePair4, txFeePair5, txFeePair6} //nolint: gochecknoglobals // no need to lint this global variable

func TestMiner_MineBlock(t *testing.T) {
	mempool := NewMemPool()
	for _, v := range txs {
		mempool.AppendTransaction(v.Transaction, v.Fee)
	}
	miner := Miner{
		mempool:      mempool,
		hasherType:   hash.SHA256,
		minerAddress: "minerAddress",
		blockHeight:  0,
		target:       16,
	}

	// simple block mining with hash difficulty 16
	ctx := context.Background()
	block, err := miner.MineBlock(ctx)
	require.NoError(t, err)
	assert.Len(t, block.Transactions, len(txs)+1)
	assert.True(t, block.Transactions[0].IsCoinbase())
	assert.Greater(t, block.Header.Nonce, uint(0))
	assert.Equal(t, script.NewScript(script.P2PK, []byte("minerAddress")), block.Transactions[0].Vout[0].ScriptPubKey)
	assert.Equal(t, []byte{0x0, 0x0}, block.Hash[:2])

	// cancel block in the middle of mining aborting the process
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err = miner.MineBlock(ctx)
	require.Error(t, err)
}

func TestMiner_createCoinbaseTransaction(t *testing.T) {
	miner := NewMiner("minerAddress", hash.SHA256)

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
