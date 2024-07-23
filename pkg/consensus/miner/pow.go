package miner

import (
	"chainnet/pkg/consensus/util"
	"chainnet/pkg/crypto/hash"
	"chainnet/pkg/kernel"
	"context"
	"errors"
	"fmt"
	"math/big"
	"runtime"
	"sync"
)

const (
	HashLength = 256
	MaxNonce   = ^uint(0)
)

type miningResult struct {
	hash  []byte
	nonce uint
	err   error
}

// ProofOfWork holds the components needed for mining
type ProofOfWork struct {
	ctx            context.Context
	results        chan miningResult
	wg             sync.WaitGroup
	hashDifficulty *big.Int

	hasherType hash.HasherType

	bh *kernel.BlockHeader
}

// NewProofOfWork creates a new ProofOfWork instance
func NewProofOfWork(ctx context.Context, bh *kernel.BlockHeader, hasherType hash.HasherType) (*ProofOfWork, error) {
	if bh.Target >= HashLength {
		return nil, errors.New("target is bigger than the hash length")
	}

	hashDifficulty := big.NewInt(1)
	hashDifficulty.Lsh(hashDifficulty, HashLength-bh.Target)

	return &ProofOfWork{
		ctx:            ctx,
		results:        make(chan miningResult),
		hashDifficulty: hashDifficulty,
		hasherType:     hasherType,
		bh:             bh,
	}, nil
}

// CalculateBlockHash calculates the hash of a block in parallel
func (pow *ProofOfWork) CalculateBlockHash() ([]byte, uint, error) {
	if pow.bh.Target >= HashLength {
		return nil, 0, errors.New("target is bigger than the hash length")
	}

	numGoroutines := runtime.NumCPU()
	nonceRange := MaxNonce / uint(numGoroutines)

	// split work and ranges among go routines
	for i := range make([]int, numGoroutines) {
		pow.wg.Add(1)
		go pow.startMining(pow.bh, nonceRange, uint(i)*nonceRange)
	}

	// wait for all go routines to finish
	go func() {
		pow.wg.Wait()
		close(pow.results)
	}()

	// wait for the first result to be returned
	for result := range pow.results {
		if result.err == nil {
			return result.hash, result.nonce, nil
		} else if result.err.Error() == "mining cancelled by context" {
			return nil, 0, result.err
		}
	}

	return nil, 0, errors.New("could not find hash for kernel")
}

// startMining starts a mining process in a goroutine
func (pow *ProofOfWork) startMining(bh *kernel.BlockHeader, nonceRange uint, startNonce uint) {
	defer pow.wg.Done()
	// initialize hash function in each goroutine because is not thread safe by default
	hasher := hash.GetHasher(pow.hasherType)
	var localHashInt big.Int

	// iterate over the nonce range and calculate the hash
	for nonce := startNonce; nonce < startNonce+nonceRange && nonce < MaxNonce; nonce++ {
		select {
		case <-pow.ctx.Done():
			// if the context is cancelled, return immediately
			pow.results <- miningResult{nil, 0, errors.New("mining cancelled by context")}
			return
		default:
			// calculate the hash of the block
			bh.SetNonce(nonce)
			blockHash, err := util.CalculateBlockHash(bh, hasher)
			if err != nil {
				pow.results <- miningResult{nil, 0, fmt.Errorf("could not hash block: %w", err)}
				return
			}

			localHashInt.SetBytes(blockHash)

			// check if the hash meets the difficulty
			if localHashInt.Cmp(pow.hashDifficulty) == -1 {
				pow.results <- miningResult{blockHash, nonce, nil}
				return
			}
		}
	}
}
