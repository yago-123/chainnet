package miner

import (
	"chainnet/pkg/kernel"
	"fmt"
)

type Miner struct {
	mempool MemPool
}

func NewMiner() *Miner {
	return &Miner{
		mempool: NewMemPool(),
	}
}

func (m *Miner) MineBlock() (*kernel.Block, error) {

	m.collectTransactions()

	m.createCoinbaseTransaction()

	m.createBlockHeader()

	m.calculateHash()

	// todo(): add as different method?
	m.broadcastBlock()

	return nil, fmt.Errorf("unable to mine block")
}

func (m *Miner) collectTransactions() ([]*kernel.Transaction, error) {
	return []*kernel.Transaction{}, nil
}

func (m *Miner) createCoinbaseTransaction() *kernel.Transaction {
	return &kernel.Transaction{}
}

func (m *Miner) createBlockHeader() *kernel.BlockHeader {
	return &kernel.BlockHeader{}
}

func (m *Miner) calculateHash() {

}

func (m *Miner) broadcastBlock() {

}
