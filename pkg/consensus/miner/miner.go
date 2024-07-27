package miner

import (
	blockchain "chainnet/pkg/chain"
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
	// AdjustDifficultyHeight adjusts difficulty every 2016 blocks (~2 weeks)
	AdjustDifficultyHeight = 2016

	BlockVersion = "0.0.1"

	MinerObserverID = "miner"
)

type Miner struct {
	mempool MemPool
	// import hasher type instead of directly hasher because will be used in multi-threaded scenario
	hasherType hash.HasherType
	chain      *blockchain.Blockchain

	minerAddress []byte
	blockHeight  uint
	target       uint

	isMining bool
	ctx      context.Context
	cancel   context.CancelFunc
}

func NewMiner(publicKey []byte, hasherType hash.HasherType, chain *blockchain.Blockchain) *Miner {
	return &Miner{
		mempool:      NewMemPool(),
		hasherType:   hasherType,
		chain:        chain,
		minerAddress: publicKey,
		blockHeight:  0,
		target:       1,
	}
}

func (m *Miner) AdjustMiningDifficulty() uint {
	// todo(): implement mining difficulty adjustment
	return AdjustDifficultyHeight
}

func (m *Miner) CancelMining() {
	// todo(): take into account the block height before canceling
	if m.isMining {
		m.cancel()
		m.isMining = false
	}
}

// MineBlock assemble, generates and mines a new block
func (m *Miner) MineBlock() (*kernel.Block, error) {
	var err error

	if m.isMining {
		// impossible case in theory
		return nil, fmt.Errorf("miner is already mining ")
	}

	// create context for canceling mining if needed (check OnBlockAddition observer func)
	m.isMining = true
	m.ctx, m.cancel = context.WithCancel(context.Background())

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
		case <-m.ctx.Done():
			// abort mining if the context is cancelled
			return nil, fmt.Errorf("mining cancelled by context")
		default:
			// start mining the block (proof of work)
			pow, errPow := NewProofOfWork(m.ctx, blockHeader, m.hasherType)
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

			// add the block to the chain
			if err := m.chain.AddBlock(block); err != nil {
				return nil, fmt.Errorf("unable to add block to the chain: %w", err)
			}

			return block, nil
		}
	}
}

// ID returns the observer id
func (m *Miner) ID() string {
	return MinerObserverID
}

// OnBlockAddition is called when a new block is added to the blockchain via the observer pattern
func (m *Miner) OnBlockAddition(_ *kernel.Block) {
	// cancel previous mining
	m.CancelMining()

	// start new mining process retrieving the last height specifically from the blockchain itself, the value
	// is queried directly in order to ensure that we get the latest value (as opposed to retrieving from storage,
	// given that it may have not been written yet)
	m.blockHeight = m.chain.GetLastHeight() + 1
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
	tx := kernel.NewCoinbaseTransaction(string(m.minerAddress), reward, collectedFee)
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
