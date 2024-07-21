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
	CoinbaseReward              = 50
	NumberOfTransactionsInBlock = 10
)

type Miner struct {
	mempool MemPool
	hasher  hash.Hashing
	pow     *ProofOfWork
}

func NewMiner() *Miner {
	return &Miner{
		mempool: NewMemPool(),
		hasher:  hash.NewSHA256(),
		pow:     NewProofOfWork(256, hash.NewSHA256()),
	}
}

// MineBlock assemble, generates and mines a new block
func (m *Miner) MineBlock(ctx context.Context) (*kernel.Block, error) {
	// retrieve transactions that are going to be placed inside the block
	collectedTxs, collectedFee, err := m.collectTransactions()
	if err != nil {
		return nil, fmt.Errorf("unable to collect transactions from mempool: %v", err)
	}

	// generate the coinbase transaction and add to the list of transactions
	coinbaseTx := m.createCoinbaseTransaction(collectedFee)
	txs := append([]*kernel.Transaction{coinbaseTx}, collectedTxs...)

	// create block header
	blockHeader, err := m.createBlockHeader(txs, 0, []byte("prevBlockHash"), 1)
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
			blockHash, nonce, err := m.pow.CalculateBlockHash(blockHeader, ctx)
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

func (m *Miner) collectTransactions() ([]*kernel.Transaction, uint, error) {
	txs := []*kernel.Transaction{}

	txs, totalFee := m.mempool.RetrieveTransactions(NumberOfTransactionsInBlock)

	// todo(): check that there are no conflictive transactions retrieved

	return txs, totalFee, nil
}

func (m *Miner) createCoinbaseTransaction(collectedFee uint) *kernel.Transaction {
	// todo(): make coinbase reward variable based on height of the blockchain (halving)
	return kernel.NewCoinbaseTransaction("miner", CoinbaseReward, collectedFee)
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
