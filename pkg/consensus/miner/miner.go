package miner

import (
	"chainnet/pkg/consensus"
	"chainnet/pkg/consensus/util"
	"chainnet/pkg/crypto/hash"
	"chainnet/pkg/kernel"
	"context"
	"fmt"
	"time"
)

const (
	InitialCoinbaseReward = 50
	HalvingInterval       = 210000
	MaxNumberHalvings     = 64

	BlockVersion = "0.0.1"

	MinerObserverId = "miner"
)

type Miner struct {
	mempool MemPool
	// import hasher type instead of directly hasher because will be used in multi-threaded scenario
	hasherType   hash.HasherType
	minerAddress string
	blockHeight  uint
	target       uint
}

func NewMiner(minerAddress string, hasherType hash.HasherType) *Miner {
	return &Miner{
		mempool:      NewMemPool(),
		hasherType:   hasherType,
		minerAddress: minerAddress,
		blockHeight:  0,
		target:       1,
	}
}

func (m *Miner) AdjustMiningDifficulty() {
	// todo(): implement mining difficulty adjustment
}

// MineBlock assemble, generates and mines a new block
func (m *Miner) MineBlock(ctx context.Context) (*kernel.Block, error) {
	var err error

	// retrieve transactions that are going to be placed inside the block
	collectedTxs, collectedFee := m.mempool.RetrieveTransactions(kernel.MaxNumberTxsPerBlock)

	// generate the coinbase transaction and add to the list of transactions
	coinbaseTx, err := m.createCoinbaseTransaction(collectedFee, m.blockHeight)
	if err != nil {
		return nil, fmt.Errorf("unable to create coinbase transaction: %w", err)
	}
	txs := append([]*kernel.Transaction{coinbaseTx}, collectedTxs...)

	// todo(): handle prevBlockHash and block height
	// create block header
	blockHeader, err := m.createBlockHeader(txs, m.blockHeight, []byte("prevBlockHash"), m.target)
	if err != nil {
		return nil, fmt.Errorf("unable to create block header: %w", err)
	}

	// start mining process
	for {
		select {
		case <-ctx.Done():
			// abort mining if the context is cancelled
			return nil, fmt.Errorf("mining cancelled by context")
		default:
			// start mining the block (proof of work)
			pow, errPow := NewProofOfWork(ctx, blockHeader, m.hasherType)
			if errPow != nil {
				return nil, fmt.Errorf("unable to create proof of work: %w", errPow)
			}
			blockHash, nonce, errPow := pow.CalculateBlockHash()
			if errPow != nil {
				// if no nonce was found, readjust the timestamp and try again
				blockHeader.SetTimestamp(time.Now().Unix())
				continue
			}

			// assemble the whole block and return it
			blockHeader.SetNonce(nonce)
			block := kernel.NewBlock(blockHeader, txs, blockHash)

			// todo(): validate block before returning it?

			return block, nil
		}
	}
}

// Id returns the observer id
func (m *Miner) Id() string {
	return MinerObserverId
}

// OnBlockAddition is called when a new block is added to the blockchain via the observer pattern
func (m *Miner) OnBlockAddition(block *kernel.Block) {

}

// createCoinbaseTransaction creates a new coinbase transaction with the reward and collected fees
func (m *Miner) createCoinbaseTransaction(collectedFee, height uint) (*kernel.Transaction, error) {
	reward := uint(0)
	// calculate reward based on block height and halving interval. If height greater than 64 halvings, reward is 0
	// to avoid dealing with bugs
	halvings := height / HalvingInterval
	if halvings < MaxNumberHalvings {
		reward = uint(InitialCoinbaseReward >> halvings)
	}

	// creates transaction and calculate hash
	tx := kernel.NewCoinbaseTransaction(m.minerAddress, reward, collectedFee)
	txHash, err := util.CalculateTxHash(tx, hash.GetHasher(m.hasherType))
	if err != nil {
		return nil, fmt.Errorf("unable to calculate transaction hash: %w", err)
	}
	tx.SetID(txHash)

	return tx, nil
}

func (m *Miner) createBlockHeader(txs []*kernel.Transaction, height uint, prevBlockHash []byte, target uint) (*kernel.BlockHeader, error) {
	merkleTree, err := consensus.NewMerkleTreeFromTxs(txs, hash.GetHasher(m.hasherType))
	if err != nil {
		return nil, fmt.Errorf("unable to create Merkle tree from transactions: %w", err)
	}

	return kernel.NewBlockHeader(
		[]byte(BlockVersion),
		time.Now().Unix(),
		merkleTree.RootHash(),
		height,
		prevBlockHash,
		target,
		0,
	), nil
}
