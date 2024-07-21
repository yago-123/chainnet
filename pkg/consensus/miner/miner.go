package miner

import (
	"chainnet/pkg/consensus"
	"chainnet/pkg/crypto/hash"
	"chainnet/pkg/kernel"
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
}

func NewMiner() *Miner {
	return &Miner{
		mempool: NewMemPool(),
		hasher:  hash.NewSHA256(),
	}
}

// MineBlock assemble, generates and mines a new block
func (m *Miner) MineBlock() (*kernel.Block, error) {
	collectedTxs, collectedFee, err := m.collectTransactions()
	if err != nil {
		return nil, fmt.Errorf("unable to collect transactions from mempool: %v", err)
	}

	coinbaseTx := m.createCoinbaseTransaction(collectedFee)

	txs := append([]*kernel.Transaction{coinbaseTx}, collectedTxs...)

	blockHeader, err := m.createBlockHeader(txs, 0, []byte("prevBlockHash"))
	if err != nil {
		return nil, fmt.Errorf("unable to create block header: %v", err)
	}

	_ = kernel.NewBlock(blockHeader, txs)

	m.calculateHash()

	// todo(): add as different method?
	m.broadcastBlock()

	return nil, fmt.Errorf("unable to mine block")
}

func (m *Miner) collectTransactions() ([]*kernel.Transaction, uint, error) {
	txs := []*kernel.Transaction{}

	memPoolSize := m.mempool.Len()

	// return error if MemPool is empty
	if memPoolSize == 0 {
		return []*kernel.Transaction{}, 0, fmt.Errorf("no transactions in mempool to mine block")
	}

	// specify the number of transactions to retrieve from MemPool
	numTxsRetrieve := NumberOfTransactionsInBlock
	if numTxsRetrieve > memPoolSize {
		numTxsRetrieve = m.mempool.Len()
	}

	// retrieve transactions from MemPool
	// todo(): adjust so is not fixed size and other variables are taken into account (size, fee, etc)
	totalFee := uint(0)
	for _ = range numTxsRetrieve {
		tx, fee := m.mempool.Pop()
		if tx == nil {
			continue
		}

		// sum fee and collect txs
		totalFee += fee
		txs = append(txs, tx)
	}

	// if no transactions were retrieved (txs are validated when entering MemPool)
	// in theory this scenario should be impossible, txs must be already validated before added to MemPool
	if len(txs) == 0 {
		return []*kernel.Transaction{}, 0, fmt.Errorf("no valid transactions were available")
	}

	return txs, totalFee, nil
}

func (m *Miner) createCoinbaseTransaction(collectedFee uint) *kernel.Transaction {
	// todo(): make coinbase reward variable based on height of the blockchain (halving)
	return kernel.NewCoinbaseTransaction("miner", CoinbaseReward, collectedFee)
}

func (m *Miner) createBlockHeader(txs []*kernel.Transaction, height uint, prevBlockHash []byte) (*kernel.BlockHeader, error) {
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
		0,
		0,
	), nil
}

func (m *Miner) calculateHash() {

}

func (m *Miner) broadcastBlock() {

}
