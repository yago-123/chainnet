package miner //nolint:testpackage // don't create separate package for tests

import (
	"chainnet/config"
	blockchain "chainnet/pkg/chain"
	"chainnet/pkg/chain/observer"
	"chainnet/pkg/consensus/util"
	"chainnet/pkg/crypto/hash"
	"chainnet/pkg/kernel"
	"chainnet/pkg/script"
	"chainnet/tests/mocks/consensus"
	mockStorage "chainnet/tests/mocks/storage"
	"context"
	"testing"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var txFeePair1 = txFeePair{ //nolint: gochecknoglobals // no need to lint this global variable
	Transaction: &kernel.Transaction{
		ID:   []byte("id1"),
		Vin:  []kernel.TxInput{kernel.NewInput([]byte("txid"), 0, "scriptSig", "pubKey")},
		Vout: []kernel.TxOutput{kernel.NewOutput(1, script.P2PK, "scriptPubKey")},
	},
	Fee: 10,
}

var txFeePair2 = txFeePair{ //nolint: gochecknoglobals // no need to lint this global variable
	Transaction: &kernel.Transaction{
		ID:   []byte("id2"),
		Vin:  []kernel.TxInput{kernel.NewInput([]byte("txid2"), 1, "scriptSig", "pubKey")},
		Vout: []kernel.TxOutput{kernel.NewOutput(2, script.P2PK, "scriptPubKey")},
	},
	Fee: 2,
}

var txFeePair3 = txFeePair{ //nolint: gochecknoglobals // no need to lint this global variable
	Transaction: &kernel.Transaction{
		ID:   []byte("id3"),
		Vin:  []kernel.TxInput{kernel.NewInput([]byte("txid3"), 2, "scriptSig", "pubKey")},
		Vout: []kernel.TxOutput{kernel.NewOutput(3, script.P2PK, "scriptPubKey")},
	},
}

var txFeePair4 = txFeePair{ //nolint: gochecknoglobals // no need to lint this global variable
	Transaction: &kernel.Transaction{
		ID:   []byte("id4"),
		Vin:  []kernel.TxInput{kernel.NewInput([]byte("txid4"), 3, "scriptSig", "pubKey")},
		Vout: []kernel.TxOutput{kernel.NewOutput(4, script.P2PK, "scriptPubKey")},
	},
	Fee: 1,
}

var txFeePair5 = txFeePair{ //nolint: gochecknoglobals // no need to lint this global variable
	Transaction: &kernel.Transaction{
		ID:   []byte("id5"),
		Vin:  []kernel.TxInput{kernel.NewInput([]byte("txid5"), 4, "scriptSig", "pubKey")},
		Vout: []kernel.TxOutput{kernel.NewOutput(5, script.P2PK, "scriptPubKey")},
	},
	Fee: 9,
}

var txFeePair6 = txFeePair{ //nolint: gochecknoglobals // no need to lint this global variable
	Transaction: &kernel.Transaction{
		ID:   []byte("id6"),
		Vin:  []kernel.TxInput{kernel.NewInput([]byte("txid6"), 5, "scriptSig", "pubKey")},
		Vout: []kernel.TxOutput{kernel.NewOutput(6, script.P2PK, "scriptPubKey")},
	},
	Fee: 6,
}

var txs = []txFeePair{txFeePair1, txFeePair2, txFeePair3, txFeePair4, txFeePair5, txFeePair6} //nolint: gochecknoglobals // no need to lint this global variable

func TestMiner_MineBlock(t *testing.T) {
	mempool := NewMemPool()
	for _, v := range txs {
		txID, err := util.CalculateTxHash(v.Transaction, hash.NewSHA256())
		require.NoError(t, err)

		v.Transaction.SetID(txID)
		mempool.AppendTransaction(v.Transaction, v.Fee)
	}

	storage := &mockStorage.MockStorage{}
	storage.
		On("GetLastHeader").
		Return(&kernel.BlockHeader{}, nil)
	storage.
		On("GetLastBlockHash").
		Return([]byte{}, nil)

	chain, err := blockchain.NewBlockchain(&config.Config{Logger: logrus.New()}, storage, hash.NewSHA256(), consensus.NewMockHeavyValidator(), observer.NewSubjectObserver())
	require.NoError(t, err)

	miner := Miner{
		mempool:      mempool,
		hasherType:   hash.SHA256,
		minerAddress: []byte("minerAddress"),
		target:       16,
		chain:        chain,
	}

	// simple block mining with hash difficulty 16
	block, err := miner.MineBlock()
	require.NoError(t, err)
	assert.Len(t, block.Transactions, len(txs)+1)
	assert.True(t, block.Transactions[0].IsCoinbase())
	assert.Greater(t, block.Header.Nonce, uint(0))
	assert.Equal(t, script.NewScript(script.P2PK, []byte("minerAddress")), block.Transactions[0].Vout[0].ScriptPubKey)
	assert.Equal(t, []byte{0x0, 0x0}, block.Hash[:2])

	// cancel block in the middle of mining aborting the process
	miner.ctx, miner.cancel = context.WithCancel(context.Background())
	miner.isMining = true
	miner.cancel()
	_, err = miner.MineBlock()
	require.Error(t, err)
}

func TestMiner_createCoinbaseTransaction(t *testing.T) {
	storage := &mockStorage.MockStorage{}
	storage.
		On("GetLastHeader").
		Return(&kernel.BlockHeader{}, nil)
	storage.
		On("GetLastBlockHash").
		Return([]byte{}, nil)

	chain, err := blockchain.NewBlockchain(&config.Config{Logger: logrus.New()}, storage, hash.NewSHA256(), consensus.NewMockHeavyValidator(), observer.NewSubjectObserver())
	require.NoError(t, err)
	miner := NewMiner(config.NewConfig(logrus.New(), time.Second*10), []byte("minerAddress"), chain, NewMemPool(), hash.SHA256)

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
