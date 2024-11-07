package util_test

import (
	util_p2pkh "github.com/yago-123/chainnet/pkg/util/p2pkh"
	"testing"

	"github.com/btcsuite/btcutil/base58"
	"github.com/yago-123/chainnet/pkg/util"

	"github.com/stretchr/testify/assert"

	"github.com/stretchr/testify/require"
)

func TestIsFirstNBytesZero(t *testing.T) {
	hash := []byte{0x0, 0xFF, 0xFF}

	require.True(t, util.IsFirstNBitsZero(hash, 8))
	require.False(t, util.IsFirstNBitsZero(hash, 16))
	require.False(t, util.IsFirstNBitsZero(hash, 256))

	hash = []byte{0x7F, 0xFF, 0xFF}
	require.True(t, util.IsFirstNBitsZero(hash, 1))
	require.False(t, util.IsFirstNBitsZero(hash, 2))

	hash = []byte{0x0, 0x7F, 0xFF}
	require.True(t, util.IsFirstNBitsZero(hash, 9))
	require.False(t, util.IsFirstNBitsZero(hash, 10))
}

func TestCalculateMiningDifficulty(t *testing.T) {
	type args struct {
		currentTarget  uint
		targetTimeSpan float64
		actualTimeSpan int64
	}
	tests := []struct {
		name string
		args args
		want uint
	}{
		{
			name: "No adjustment needed",
			args: args{
				currentTarget:  100,
				targetTimeSpan: 600,
				actualTimeSpan: 600,
			},
			want: 100,
		},
		{
			name: "Increase difficulty (actual time is less than target time)",
			args: args{
				currentTarget:  100,
				targetTimeSpan: 600,
				actualTimeSpan: 300,
			},
			want: 101,
		},
		{
			name: "Decrease difficulty",
			args: args{
				currentTarget:  100,
				targetTimeSpan: 600,
				actualTimeSpan: 1200,
			},
			want: 99,
		},
		{
			name: "Decrease difficulty by a small margin ",
			args: args{
				currentTarget:  100,
				targetTimeSpan: 600,
				actualTimeSpan: 601,
			},
			want: 99,
		},
		{
			name: "Test lower limits",
			args: args{
				currentTarget:  1,
				targetTimeSpan: 600,
				actualTimeSpan: 700, // twice the target time span
			},
			want: 1,
		},
		{
			name: "Test upper limits",
			args: args{
				currentTarget:  255,
				targetTimeSpan: 600,
				actualTimeSpan: 500, // twice the target time span
			},
			want: 255,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := util.CalculateMiningTarget(tt.args.currentTarget, tt.args.targetTimeSpan, tt.args.actualTimeSpan); got != tt.want {
				t.Errorf("CalculateMiningDifficulty() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsValidHash(t *testing.T) {
	hash := "0000006484ffdc39a5ba6cebae9e398878f24bcab93f4c32acf81e246fa2474b"
	assert.True(t, util.IsValidHash([]byte(hash)))
}

func TestGenerateP2PKHAddrFromPubKey(t *testing.T) {
	p2pkhAddr, err := util_p2pkh.GenerateP2PKHAddrFromPubKey(
		base58.Decode("aSq9DsNNvGhYxYyqA9wd2eduEAZ5AXWgJTbTJdddpT9aV3HbEPRuBpyEXFktCPCgrdp3FEXrfqjz2xoeQwTCqBs8qJtUFNmCLRTyVaTYuy7G8RZnHkABrMpH2cCG"),
		1,
	)

	require.NoError(t, err)
	require.Len(t, string(p2pkhAddr), util_p2pkh.P2PKHAddressLength)
	assert.Equal(t, "agr72ArMnsmdm9XTScgCpXnwkhAANyBCd", base58.Encode(p2pkhAddr))
}

func TestExtractPubKeyHashedFromP2PKHAddr(t *testing.T) {
	pubKeyHash, version, err := util_p2pkh.ExtractPubKeyHashedFromP2PKHAddr(
		base58.Decode("agr72ArMnsmdm9XTScgCpXnwkhAANyBCd"),
	)

	require.NoError(t, err)
	assert.Equal(t, 1, int(version))
	assert.Equal(t, "2ajHyKQLikZqXV9rpaSfnV6mh7a5", base58.Encode(pubKeyHash))
	assert.Len(t, pubKeyHash, 20)

	// modify byte to test checksum validation
	_, _, err = util_p2pkh.ExtractPubKeyHashedFromP2PKHAddr(
		base58.Decode("agr72ArMnsmd99XTScgCpXnwkhAANyBCd"),
	)

	require.Error(t, err)

	// make sure that length is checked
	_, _, err = util_p2pkh.ExtractPubKeyHashedFromP2PKHAddr(
		base58.Decode("agr72ArMnsmd9XTScgCpXnwkhAANyBCd"),
	)
	require.Error(t, err)
}
