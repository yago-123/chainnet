package miner //nolint:testpackage // don't create separate package for tests

import (
	"chainnet/pkg/crypto/hash"
	"chainnet/pkg/kernel"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProofOfWork_CalculateBlockHash(t *testing.T) {
	ctx := context.Background()

	// check that returns error if the target is bigger than the hash itself
	bh := kernel.NewBlockHeader([]byte("1"), 1, []byte("merkle-root"), 1, []byte("prev-block-hash"), 300, 0)
	_, err := NewProofOfWork(ctx, bh, hash.SHA256)
	require.Error(t, err)

	// calculate simple hash with 1 zero
	bh = kernel.NewBlockHeader([]byte("1"), 1, []byte("merkle-root"), 1, []byte("prev-block-hash"), 8, 0)
	pow, err := NewProofOfWork(ctx, bh, hash.SHA256)
	require.NoError(t, err)
	blockHash, nonce, err := pow.CalculateBlockHash()
	require.NoError(t, err)
	assert.Positive(t, nonce)
	assert.Equal(t, []byte{0x0}, blockHash[:1])
	assert.NotEqual(t, []byte{0x0}, blockHash[1:2])

	// calculate simple hash with 2 zeros
	bh = kernel.NewBlockHeader([]byte("1"), 1, []byte("merkle-root"), 1, []byte("prev-block-hash"), 16, 0)
	pow, err = NewProofOfWork(ctx, bh, hash.SHA256)
	require.NoError(t, err)
	blockHash, nonce, err = pow.CalculateBlockHash()
	require.NoError(t, err)
	assert.Positive(t, nonce)
	assert.Equal(t, []byte{0x0, 0x0}, blockHash[:2])
	assert.NotEqual(t, []byte{0x0}, blockHash[2:3])

	// make suire that proof of work can be cancelled
	bh = kernel.NewBlockHeader([]byte("1"), 1, []byte("merkle-root"), 1, []byte("prev-block-hash"), 200, 0)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	pow, err = NewProofOfWork(ctx, bh, hash.SHA256)
	require.NoError(t, err)
	_, _, err = pow.CalculateBlockHash()
	require.Error(t, err)
}
