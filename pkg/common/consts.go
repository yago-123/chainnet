package common

import "github.com/yago-123/chainnet/pkg/kernel"

const (
	InitialCoinbaseReward = 50 * kernel.ChainnetCoinAmount
	HalvingInterval       = 210000
	MaxNumberHalvings     = 64
)
