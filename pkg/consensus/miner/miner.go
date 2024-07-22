package miner

import (
	"chainnet/pkg/consensus"
	"chainnet/pkg/crypto/hash"
	"chainnet/pkg/kernel"
	"context"
	"fmt"
	"time"
)

const (
	InitialCoinbaseReward       = 50
	HalvingInterval             = 210000
	NumberOfTransactionsInBlock = 10
)

type Miner struct {
	mempool      MemPool
	hasher       hash.Hashing
	minerAddress string
	blockHeight  uint
	target       uint
}

func NewMiner(minerAddress string) *Miner {
	return &Miner{
		mempool:      NewMemPool(),
		hasher:       hash.NewSHA256(),
		minerAddress: minerAddress,
		blockHeight:  0,
		target:       1,
	}
}

// MineBlock assemble, generates and mines a new block
func (m *Miner) MineBlock(ctx context.Context) (*kernel.Block, error) {
	var err error

	// retrieve transactions that are going to be placed inside the block
	collectedTxs, collectedFee := m.mempool.RetrieveTransactions(NumberOfTransactionsInBlock)

	// generate the coinbase transaction and add to the list of transactions
	coinbaseTx := m.createCoinbaseTransaction(collectedFee, m.blockHeight)
	txs := append([]*kernel.Transaction{coinbaseTx}, collectedTxs...)

	// create block header
	blockHeader, err := m.createBlockHeader(txs, m.blockHeight, []byte("prevBlockHash"), m.target)
	if err != nil {
		return nil, fmt.Errorf("unable to create block header: %v", err)
	}

	// start mining process
	for {
		select {
		case <-ctx.Done():
			// abort mining if the context is cancelled
			return nil, fmt.Errorf("mining cancelled by context")
		default:
			// start mining the block (proof of work)
			pow, err := NewProofOfWork(ctx, blockHeader)
			if err != nil {
				return nil, fmt.Errorf("unable to create proof of work: %v", err)
			}
			blockHash, nonce, err := pow.CalculateBlockHash()
			if err != nil {
				// if no nonce was found, readjust the timestamp and try again
				blockHeader.SetTimestamp(time.Now().Unix())
				continue
			}

			// assemble the whole block and return it
			blockHeader.SetNonce(nonce)
			block := kernel.NewBlock(blockHeader, txs, blockHash)

			return block, nil
		}
	}
}

// createCoinbaseTransaction creates a new coinbase transaction with the reward and collected fees
func (m *Miner) createCoinbaseTransaction(collectedFee, height uint) *kernel.Transaction {
	// calculate the current reward based on block height
	halvings := height / HalvingInterval
	reward := uint(InitialCoinbaseReward >> halvings)

	return kernel.NewCoinbaseTransaction(m.minerAddress, reward, collectedFee)
}

func (m *Miner) createBlockHeader(txs []*kernel.Transaction, height uint, prevBlockHash []byte, target uint) (*kernel.BlockHeader, error) {
	merkleTree, err := consensus.NewMerkleTreeFromTxs(txs, m.hasher)
	if err != nil {
		return nil, fmt.Errorf("unable to create Merkle tree from transactions: %v", err)
	}

	return kernel.NewBlockHeader(
		[]byte("0.0.1"),
		time.Now().Unix(),
		merkleTree.RootHash(),
		height,
		prevBlockHash,
		target,
		0,
	), nil
}
