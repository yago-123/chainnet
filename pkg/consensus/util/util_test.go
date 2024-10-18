package util //nolint:testpackage // don't create separate package for tests

import (
	"testing"

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
		currentDifficulty float64
		targetTimeSpan    float64
		actualTimeSpan    int64
	}
	tests := []struct {
		name string
		args args
		want float64
	}{
		{
			name: "No adjustment needed",
			args: args{
				currentDifficulty: 100.0,
				targetTimeSpan:    600,
				actualTimeSpan:    600,
			},
			want: 100.0,
		},
		{
			name: "Increase difficulty (actual time is less than target time)",
			args: args{
				currentDifficulty: 100.0,
				targetTimeSpan:    600,
				actualTimeSpan:    300, // half the target time span
			},
			want: 50.0, // 100 * (300 / 600)
		},
		{
			name: "Decrease difficulty (actual time is more than target time)",
			args: args{
				currentDifficulty: 100.0,
				targetTimeSpan:    600,
				actualTimeSpan:    1200, // twice the target time span
			},
			want: 200.0, // 100 * (1200 / 600)
		},
		{
			name: "Limit increase to 4x adjustment",
			args: args{
				currentDifficulty: 100.0,
				targetTimeSpan:    600,
				actualTimeSpan:    100, // less than half the target time span
			},
			want: 25.0, // 100 * (100 / 600) would be less than 4x
		},
		{
			name: "Limit decrease to 1/4x adjustment",
			args: args{
				currentDifficulty: 100.0,
				targetTimeSpan:    600,
				actualTimeSpan:    2400, // four times the target time span
			},
			want: 400.0, // should be limited to 100 / 4
		},
		{
			name: "Exact adjustment within limits",
			args: args{
				currentDifficulty: 80.0,
				targetTimeSpan:    800,
				actualTimeSpan:    400, // half the target time span
			},
			want: 40.0, // 80 * (400 / 800)
		},
		{
			name: "Exact adjustment within limits",
			args: args{
				currentDifficulty: 80.0,
				targetTimeSpan:    800,
				actualTimeSpan:    400,
			},
			want: 40.0,
		},
		{
			name: "Zero currentDifficulty",
			args: args{
				currentDifficulty: 0.0,
				targetTimeSpan:    600,
				actualTimeSpan:    300,
			},
			want: InitialDifficulty,
		},
		{
			name: "Zero targetTimeSpan",
			args: args{
				currentDifficulty: 100.0,
				targetTimeSpan:    0.0,
				actualTimeSpan:    600,
			},
			want: 400.0, // current difficulty multiplied by MaxTargetAjustmentFactor
		},
		{
			name: "Zero actualTimeSpan",
			args: args{
				currentDifficulty: 100.0,
				targetTimeSpan:    600,
				actualTimeSpan:    0,
			},
			want: 25.0, // current difficulty divided by MaxTargetAjustmentFactor
		},
		{
			name: "Zero targetTimeSpan and actualTimeSpan",
			args: args{
				currentDifficulty: 100.0,
				targetTimeSpan:    0.0,
				actualTimeSpan:    0,
			},
			want: 400.0, // current difficulty multiplied by MaxTargetAjustmentFactor
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CalculateMiningDifficulty(tt.args.currentDifficulty, tt.args.targetTimeSpan, tt.args.actualTimeSpan); got != tt.want {
				t.Errorf("CalculateMiningDifficulty() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCalculateTargetFromDifficulty(t *testing.T) {
	type args struct {
		difficulty float64
	}
	tests := []struct {
		name string
		args args
		want uint
	}{
		{
			name: "Minimum difficulty",
			args: args{difficulty: 0.0},
			want: 0, // 0 * 256 = 0
		},
		{
			name: "Normal difficulty below max",
			args: args{difficulty: 0.5},
			want: 128, // 0.5 * 256 = 128
		},
		{
			name: "Maximum difficulty",
			args: args{difficulty: 1.0},
			want: 256, // 1.0 * 256 = 256, assuming MaxHashZeros is 256
		},
		{
			name: "Exceeding maximum difficulty",
			args: args{difficulty: 1.5},
			want: 256, // capped at MaxHashZeros
		},
		{
			name: "Negative difficulty",
			args: args{difficulty: -0.5},
			want: InitialDifficulty, // negative difficulty should result in 0 leading zeros
		},
		{
			name: "Difficulty as float",
			args: args{difficulty: 0.75},
			want: 192, // 0.75 * 256 = 192
		},
		{
			name: "Slightly below max difficulty",
			args: args{difficulty: 0.99},
			want: 253, // 0.99 * 256 = 253,44, rounded down to 253
		},
		{
			name: "Very high difficulty",
			args: args{difficulty: 10.0},
			want: 256, // capped at MaxHashZeros
		},
		{
			name: "Exact fraction",
			args: args{difficulty: 0.25},
			want: 64, // 0.25 * 256 = 64
		},
		{
			name: "Fraction leading to non-integer",
			args: args{difficulty: 0.1},
			want: 25, // 0.1 * 256 = 25.6, rounded down to 25
		},
		{
			name: "Extremely low difficulty",
			args: args{difficulty: 0.001},
			want: 0, // since it's so low, it results in 0
		},
		{
			name: "Moderate difficulty",
			args: args{difficulty: 0.33},
			want: 84, // 0.33 * 256 = 84.48, rounded down to 84
		},
		{
			name: "Just below 0.5 difficulty",
			args: args{difficulty: 0.49},
			want: 125, // 0.49 * 256 = 125.44, rounded down to 125
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CalculateTargetFromDifficulty(tt.args.difficulty); got != tt.want {
				t.Errorf("CalculateTargetFromDifficulty() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCalculateDifficultyFromTarget(t *testing.T) {
	type args struct {
		target uint
	}
	tests := []struct {
		name string
		args args
		want float64
	}{
		{
			name: "Minimum target",
			args: args{target: 0},
			want: 1.0, // (256 - 0) / 256 = 1.0
		},
		{
			name: "Normal target",
			args: args{target: 128},
			want: 0.5, // (256 - 128) / 256 = 0.5
		},
		{
			name: "Maximum target",
			args: args{target: 256},
			want: 0.0, // (256 - 256) / 256 = 0.0
		},
		{
			name: "Exceeding maximum target",
			args: args{target: 300},
			want: 0.0, // capped at max zeros, (256 - 256) / 256 = 0.0
		},
		{
			name: "Slightly below max target",
			args: args{target: 255},
			want: 0.00390625, // (256 - 255) / 256 = 1/256 = 0.00390625
		},
		{
			name: "Halfway target",
			args: args{target: 128},
			want: 0.5, // (256 - 128) / 256 = 0.5
		},
		{
			name: "Quarter target",
			args: args{target: 64},
			want: 0.75, // (256 - 64) / 256 = 0.75
		},
		{
			name: "Three quarters target",
			args: args{target: 192},
			want: 0.25, // (256 - 192) / 256 = 0.25
		},
		{
			name: "Difficulty from a target of 10",
			args: args{target: 10},
			want: 0.9609375, // (256 - 10) / 256 = 246 / 256 = 0,9609375
		},
		{
			name: "Very low target",
			args: args{target: 1},
			want: 0.99609375, // (256 - 1) / 256 = 0.99609375
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CalculateDifficultyFromTarget(tt.args.target); got != tt.want {
				t.Errorf("CalculateDifficultyFromTarget() = %v, want %v", got, tt.want)
			}
		})
	}
}
