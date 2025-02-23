package miner

import (
	"context"
	"errors"
	"fmt"
	"math"
	"math/big"
	"runtime"
	"sync"

	"github.com/yago-123/chainnet/pkg/crypto/hash"
	"github.com/yago-123/chainnet/pkg/kernel"
	"github.com/yago-123/chainnet/pkg/util"
)

const (
	HashLength = 256
	MaxNonce   = ^uint(0)

	CPUDivisionFactor = 2
)

var ErrMiningCancelled = errors.New("mining cancelled by context")

type miningResult struct {
	hash  []byte
	nonce uint
	err   error
}

// ProofOfWork holds the components needed for mining
type ProofOfWork struct {
	externalCtx    context.Context
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
		externalCtx:    ctx,
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

	// calculate the number of goroutines to use. We use half of the available CPUs because in our case the miner
	// will have other tasks to do (like listening for new blocks) and we don't want to starve the system (in miner only
	// scenarios, we could use all CPUs)
	numGoroutines := int(math.Max(1, float64(runtime.NumCPU()/CPUDivisionFactor)))
	nonceRange := MaxNonce / uint(numGoroutines) //nolint:gosec // int to uint is safe

	// if one of the goroutines finds a block, use this context to propagate the cancellation. This cancellation
	// is an overlap of the pow.externalCtx given that in case of block addition, the observer code will trigger
	// the cancellation too (however, this one is more specific and more responsive)
	blockMinedCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// split work and ranges among go routines
	for i := range make([]int, numGoroutines) {
		pow.wg.Add(1)
		go pow.startMining(blockMinedCtx, *pow.bh, nonceRange, uint(i)*nonceRange) //nolint:gosec // int to uint is safe
	}

	// wait for all go routines to finish
	go func() {
		pow.wg.Wait()
		close(pow.results)
	}()

	// wait for the first result to be returned
	for result := range pow.results {
		if result.err == nil {
			// retrieve block data mined
			return result.hash, result.nonce, nil
		}

		// if mining have been cancelled, return error
		if errors.Is(result.err, ErrMiningCancelled) {
			return nil, 0, result.err
		}
	}

	// at this point no go routine have found a valid block
	return nil, 0, errors.New("could not find hash for block being mined")
}

// startMining starts a mining process in a goroutine
func (pow *ProofOfWork) startMining(blockMinedCtx context.Context, bh kernel.BlockHeader, nonceRange uint, startNonce uint) {
	defer pow.wg.Done()
	// initialize hash function in each goroutine because is not thread safe by default
	hasher := hash.GetHasher(pow.hasherType)

	// iterate over the nonce range and calculate the hash
	for nonce := startNonce; nonce < startNonce+nonceRange && nonce < MaxNonce; nonce++ {
		select {
		case <-pow.externalCtx.Done():
			// if the context is cancelled, return immediately
			pow.results <- miningResult{nil, 0, ErrMiningCancelled}
			return
		case <-blockMinedCtx.Done():
			pow.results <- miningResult{nil, 0, ErrMiningCancelled}
			return
		default:
			// calculate the hash of the block
			bh.SetNonce(nonce)
			blockHash, err := util.CalculateBlockHash(&bh, hasher)
			if err != nil {
				pow.results <- miningResult{nil, 0, fmt.Errorf("did not found hash for block: %w", err)}
				return
			}

			if util.IsFirstNBitsZero(blockHash, bh.Target) {
				pow.results <- miningResult{blockHash, nonce, nil}
				return
			}
		}
	}
}
