package util //nolint:testpackage // don't create separate package for tests

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/stretchr/testify/require"
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
			if got := CalculateMiningTarget(tt.args.currentTarget, tt.args.targetTimeSpan, tt.args.actualTimeSpan); got != tt.want {
				t.Errorf("CalculateMiningDifficulty() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsValidHash(t *testing.T) {
	hash := "0000006484ffdc39a5ba6cebae9e398878f24bcab93f4c32acf81e246fa2474b"
	assert.True(t, IsValidHash([]byte(hash)))
}
