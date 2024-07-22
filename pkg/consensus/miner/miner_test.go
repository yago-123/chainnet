package miner

import (
	"chainnet/pkg/crypto/hash"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMiner_MineBlock(t *testing.T) {

}

func TestMiner_createCoinbaseTransaction(t *testing.T) {
	miner := NewMiner("minerAddress", hash.SHA256)

	coinbase := miner.createCoinbaseTransaction(0, 0)
	assert.Len(t, coinbase.Vout, 1)
	assert.Equal(t, uint(InitialCoinbaseReward), coinbase.Vout[0].Amount)

	coinbase = miner.createCoinbaseTransaction(1, 1)
	assert.Len(t, coinbase.Vout, 2)
	assert.Equal(t, uint(InitialCoinbaseReward), coinbase.Vout[0].Amount)
	assert.Equal(t, uint(1), coinbase.Vout[1].Amount)

	coinbase = miner.createCoinbaseTransaction(0, HalvingInterval)
	assert.Len(t, coinbase.Vout, 1)
	assert.Equal(t, uint(InitialCoinbaseReward/2), coinbase.Vout[0].Amount)

	coinbase = miner.createCoinbaseTransaction(0, HalvingInterval*2)
	assert.Len(t, coinbase.Vout, 1)
	assert.Equal(t, uint(InitialCoinbaseReward/4), coinbase.Vout[0].Amount)
}
