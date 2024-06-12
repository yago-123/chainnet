package main

import (
	"chainnet/config"
	"chainnet/pkg/block"
	blockchain "chainnet/pkg/chain"
	"chainnet/pkg/consensus"
	"chainnet/pkg/encoding"
	"chainnet/pkg/storage"
	"github.com/sirupsen/logrus"
	"net/http"
)

func main() {
	difficultyPoW := uint(4)
	maxNoncePoW := uint(100)
	baseURL := "localhost:8080"

	cfg := config.NewConfig(logrus.New(), difficultyPoW, maxNoncePoW, baseURL)
	// Initialize your blockchain and other components
	bolt, err := storage.NewBoltDB("_fixture/chainnet-store", "chainnet-bucket", encoding.NewGobEncoder(cfg.Logger), cfg.Logger)
	if err != nil {
		cfg.Logger.Fatalf("Failed to initialize BoltDB: %s", err)
	}

	bc := blockchain.NewBlockchain(cfg, consensus.NewProofOfWork(cfg), bolt)

	// Add blocks
	_, _ = bc.AddBlock([]*block.Transaction{block.NewCoinbaseTx("me", "data")})
	_, _ = bc.AddBlock([]*block.Transaction{})
	_, _ = bc.AddBlock([]*block.Transaction{})

	// Iterate through blocks
	iterator := bc.CreateIterator()
	for iterator.HasNext() {
		blk, err := iterator.GetNextBlock()
		if err != nil {
			cfg.Logger.Panicf("Error getting block: %s", err)
		}

		cfg.Logger.Infof("----------------------")
		cfg.Logger.Infof("Prev. hash: %x", blk.PrevBlockHash)
		cfg.Logger.Infof("Num transactions: %d", len(blk.Transactions))
		cfg.Logger.Infof("Hash: %x", blk.Hash)
		cfg.Logger.Infof("PoW: %t", consensus.NewProofOfWork(cfg).Validate(blk))
	}

	cfg.Logger.Infof("Server listening on %s", cfg.BaseURL)
	err = http.ListenAndServe(cfg.BaseURL, NewHTTPRouter(bc))
	if err != nil {
		cfg.Logger.Fatalf("Failed to start server: %s", err)
	}
}
