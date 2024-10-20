package miner

import (
	"context"
	"fmt"
	"time"

	"github.com/yago-123/chainnet/config"
	blockchain "github.com/yago-123/chainnet/pkg/chain"
	"github.com/yago-123/chainnet/pkg/chain/explorer"
	"github.com/yago-123/chainnet/pkg/consensus"
	"github.com/yago-123/chainnet/pkg/consensus/util"
	"github.com/yago-123/chainnet/pkg/crypto/hash"
	"github.com/yago-123/chainnet/pkg/kernel"

	"github.com/btcsuite/btcutil/base58"
)

const (
	InitialCoinbaseReward = 50
	HalvingInterval       = 210000
	MaxNumberHalvings     = 64

	BlockVersion = "1"

	MinerObserverID = "miner-observer"
)

type Miner struct {
	// hasher type instead of directly hasher because hash generation will be used in high multi-threaded scenario
	hasherType hash.HasherType
	chain      *blockchain.Blockchain
	explorer   *explorer.Explorer

	minerPubKey []byte

	isMining bool
	ctx      context.Context
	cancel   context.CancelFunc

	cfg *config.Config
}

func NewMiner(cfg *config.Config, chain *blockchain.Blockchain, hasherType hash.HasherType, explorer *explorer.Explorer) (*Miner, error) {
	if len(cfg.PubKey) == 0 {
		return nil, fmt.Errorf("public key not provided, check the config file")
	}

	pubKey := base58.Decode(cfg.PubKey)

	return &Miner{
		hasherType:  hasherType,
		chain:       chain,
		explorer:    explorer,
		minerPubKey: pubKey,
		isMining:    false,
		cfg:         cfg,
	}, nil
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

	defer m.CancelMining()
	if m.isMining {
		// impossible case in theory
		return nil, fmt.Errorf("miner is already mining ")
	}

	// create context for canceling mining if needed (check OnBlockAddition observer func)
	m.isMining = true
	m.ctx, m.cancel = context.WithCancel(context.Background())

	// calculate mining target (leading zeroes in block hash) for the block that is going to be mined
	target, err := m.explorer.GetMiningTarget(m.chain.GetLastHeight(), m.cfg.AdjustmentTargetInterval, m.cfg.MiningInterval)
	if err != nil {
		return nil, fmt.Errorf("unable to get mining target: %w", err)
	}

	// retrieve transactions that are going to be placed inside the block
	collectedTxs, collectedFee := m.chain.RetrieveMempoolTxs(kernel.MaxNumberTxsPerBlock)

	// generate the coinbase transaction and add to the list of transactions
	coinbaseTx, err := m.createCoinbaseTransaction(collectedFee, m.chain.GetLastHeight())
	if err != nil {
		return nil, fmt.Errorf("unable to create coinbase transaction: %w", err)
	}
	txs := append([]*kernel.Transaction{coinbaseTx}, collectedTxs...)

	// create block header
	blockHeader, err := m.createBlockHeader(txs, m.chain.GetLastHeight(), m.chain.GetLastBlockHash(), target)
	if err != nil {
		return nil, fmt.Errorf("unable to create block header: %w", err)
	}

	// start mining process
	for {
		select {
		case <-m.ctx.Done():
			// abort mining if the context is cancelled
			return nil, fmt.Errorf("cancelled by context (height = %d)", blockHeader.Height)
		default:
			// start mining the block (proof of work)
			pow, errPow := NewProofOfWork(m.ctx, blockHeader, m.hasherType)
			if errPow != nil {
				return nil, fmt.Errorf("unable to create proof of work: %w", errPow)
			}
			blockHash, nonce, errPow := pow.CalculateBlockHash()
			if errPow != nil {
				// if no nonce was found, readjust the timestamp and try again
				m.cfg.Logger.Errorf("didn't find hash matching target: %v", errPow)
				m.cfg.Logger.Debugf("updating timestamp and starting mining process again for block with height: %d", blockHeader.Height)
				blockHeader.SetTimestamp(time.Now().Unix())
				continue
			}

			// assemble the whole block and return it
			blockHeader.SetNonce(nonce)
			block := kernel.NewBlock(blockHeader, txs, blockHash)

			// add the block to the chain
			if err = m.chain.AddBlock(block); err != nil {
				return nil, fmt.Errorf("unable to add block to the chain: %w", err)
			}

			return block, nil
		}
	}
}

// NetObserverID returns the observer id
func (m *Miner) ID() string {
	return MinerObserverID
}

// OnBlockAddition is called when a new block is added to the blockchain via the observer pattern
func (m *Miner) OnBlockAddition(_ *kernel.Block) {
	// cancel previous mining
	m.CancelMining()
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
	tx := kernel.NewCoinbaseTransaction(string(m.minerPubKey), reward, collectedFee)
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
