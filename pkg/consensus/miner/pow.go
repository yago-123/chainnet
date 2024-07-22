package miner

import (
	"chainnet/pkg/crypto/hash"
	"chainnet/pkg/kernel"
	"context"
	"errors"
	"fmt"
	"math/big"
)

const (
	HashLength = 256
	MaxNonce   = ^uint(0)
)

func CalculateBlockHash(bh *kernel.BlockHeader, ctx context.Context) ([]byte, uint, error) {
	var hashInt big.Int
	hasher := hash.NewSHA256()

	hashDifficulty := big.NewInt(1)
	hashDifficulty.Lsh(hashDifficulty, HashLength-bh.Target)

	nonce := uint(0)

	for {
		select {
		case <-ctx.Done():
			return []byte{}, 0, errors.New("mining cancelled by context")
		default:
			if nonce >= MaxNonce {
				return []byte{}, 0, errors.New("could not find hash for kernel")
			}

			data := bh.AssembleWithNonce(nonce)
			blockHash, err := hasher.Hash(data)
			if err != nil {
				return []byte{}, 0, fmt.Errorf("could not hash block: %w", err)
			}

			// Convert blockHash to an integer
			hashInt.SetBytes(blockHash)

			// Compare the integer value of the blockHash to the hashDifficulty
			if hashInt.Cmp(hashDifficulty) == -1 {
				return blockHash, nonce, nil
			}

			nonce++
		}
	}
}
