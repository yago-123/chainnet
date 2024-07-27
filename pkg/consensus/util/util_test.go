package util //nolint:testpackage // don't create separate package for tests

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestIsFirstNBytesZero(t *testing.T) {
	hash := []byte{0x0, 0xFF, 0xFF}

	require.True(t, IsFirstNBitsZero(hash, 8))
	require.False(t, IsFirstNBitsZero(hash, 16))
	require.False(t, IsFirstNBitsZero(hash, 256))

	hash = []byte{0x7F, 0xFF, 0xFF}
	require.True(t, IsFirstNBitsZero(hash, 1))
	require.False(t, IsFirstNBitsZero(hash, 2))

	hash = []byte{0x0, 0x7F, 0xFF}
	require.True(t, IsFirstNBitsZero(hash, 9))
	require.False(t, IsFirstNBitsZero(hash, 10))
}
