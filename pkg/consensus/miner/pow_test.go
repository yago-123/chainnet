package miner

import (
	"chainnet/pkg/kernel"
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestProofOfWork_CalculateBlockHash(t *testing.T) {
	ctx := context.Background()

	bh := kernel.NewBlockHeader([]byte("1"), 1, []byte("merkle-root"), 1, []byte("prev-block-hash"), 8, 0)
	pow := NewProofOfWork(ctx, bh)
	hash, nonce, err := pow.CalculateBlockHash()
	assert.NoError(t, err)
	assert.True(t, nonce > 0)
	assert.Equal(t, []byte{0x0}, hash[:1])
	assert.NotEqual(t, []byte{0x0}, hash[1:2])

	bh = kernel.NewBlockHeader([]byte("1"), 1, []byte("merkle-root"), 1, []byte("prev-block-hash"), 16, 0)
	pow = NewProofOfWork(ctx, bh)
	hash, nonce, err = pow.CalculateBlockHash()
	assert.NoError(t, err)
	assert.True(t, nonce > 0)
	assert.Equal(t, []byte{0x0, 0x0}, hash[:2])
	assert.NotEqual(t, []byte{0x0}, hash[2:3])
}
