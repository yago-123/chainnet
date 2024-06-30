package main

import (
	"chainnet/config"
	blockchain "chainnet/pkg/chain"
	"chainnet/pkg/chain/explorer"
	iterator "chainnet/pkg/chain/iterator"
	"chainnet/pkg/consensus"
	"chainnet/pkg/crypto/hash"
	"chainnet/pkg/encoding"
	"chainnet/pkg/kernel"
	"chainnet/pkg/storage"
	"net/http"

	"github.com/sirupsen/logrus"
)

const (
	DifficultyPoW = 4
	MaxNoncePoW   = 100
)

func main() {
	baseURL := "localhost:8080"

	cfg := config.NewConfig(logrus.New(), DifficultyPoW, MaxNoncePoW, baseURL)
	// Initialize your blockchain and other components
	bolt, err := storage.NewBoltDB("_fixture/chainnet-store", "chainnet-bucket", encoding.NewGobEncoder(cfg.Logger), cfg.Logger)
	if err != nil {
		cfg.Logger.Errorf("Failed to initialize BoltDB: %s", err)
	}

	// create blockchain
	bc := blockchain.NewBlockchain(cfg, consensus.NewProofOfWork(cfg.DifficultyPoW, hash.NewSHA256()), bolt)

	// create tx0 and add kernel
	tx0, _ := bc.NewCoinbaseTransaction("me")
	_, err = bc.AddBlock([]*kernel.Transaction{tx0})
	if err != nil {
		cfg.Logger.Errorf("Failed to add kernel: %s", err)
	}

	// create tx1 and add kernel
	tx1, err := bc.NewTransaction("me", "you", 10)
	if err != nil {
		cfg.Logger.Errorf("Failed to create UTXO transaction: %s", err)
	}
	_, err = bc.AddBlock([]*kernel.Transaction{tx1})
	if err != nil {
		cfg.Logger.Errorf("Failed to add kernel: %s", err)
	}

	// create tx2 and add kernel
	tx2, err := bc.NewTransaction("me", "you", 20)
	if err != nil {
		cfg.Logger.Errorf("Failed to create UTXO transaction: %s", err)
	}
	_, err = bc.AddBlock([]*kernel.Transaction{tx2})
	if err != nil {
		cfg.Logger.Errorf("Failed to add kernel: %s", err)
	}

	// Iterate through blocks
	reverseIterator := iterator.NewReverseIterator(bolt)

	_ = reverseIterator.Initialize(bc.GetLastBlockHash())
	for reverseIterator.HasNext() {
		blk, err := reverseIterator.GetNextBlock()
		if err != nil {
			cfg.Logger.Panicf("Error getting kernel: %s", err)
		}

		cfg.Logger.Infof("----------------------")
		cfg.Logger.Infof("Prev. hash: %x", blk.PrevBlockHash)
		cfg.Logger.Infof("Num transactions: %d", len(blk.Transactions))
		cfg.Logger.Infof("Hash: %x", blk.Hash)
		// cfg.Logger.Infof("PoW: %t", consensus.NewProofOfWork(1, hash.NewSHA256()).ValidateBlock(blk))
	}

	cfg.Logger.Infof("Server listening on %s", cfg.BaseURL)

	explorer := explorer.NewExplorer(bolt)
	err = http.ListenAndServe(cfg.BaseURL, NewHTTPRouter(explorer)) //nolint:gosec // add timeout later

	if err != nil {
		cfg.Logger.Fatalf("Failed to start server: %s", err)
	}
}
