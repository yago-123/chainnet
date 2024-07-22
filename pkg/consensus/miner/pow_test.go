package miner

import (
	"chainnet/pkg/kernel"
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestProofOfWork_CalculateBlockHash(t *testing.T) {
	ctx := context.Background()
	bh := kernel.NewBlockHeader([]byte("1"), 1, []byte("merkle-root"), 1, []byte("prev-block-hash"), 30, 1)
	pow := NewProofOfWork(ctx, bh)
	hash, nonce, err := pow.CalculateBlockHash()
	assert.NoError(t, err)
	assert.Equal(t, []byte{}, hash)
	assert.Equal(t, uint(0), nonce)
}
