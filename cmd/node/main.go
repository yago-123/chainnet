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
		cfg.Logger.Errorf("Failed to initialize BoltDB: %s", err)
	}

	// create blockchain
	bc := blockchain.NewBlockchain(cfg, consensus.NewProofOfWork(cfg.DifficultyPoW), bolt)

	// create tx0 and add block
	tx0, _ := bc.NewCoinbaseTransaction("me", "data")
	_, err = bc.AddBlock([]*block.Transaction{tx0})
	if err != nil {
		cfg.Logger.Errorf("Failed to add block: %s", err)
	}

	// create tx1 and add block
	tx1, err := bc.NewTransaction("me", "you", 10)
	if err != nil {
		cfg.Logger.Errorf("Failed to create UTXO transaction: %s", err)
	}
	_, err = bc.AddBlock([]*block.Transaction{tx1})
	if err != nil {
		cfg.Logger.Errorf("Failed to add block: %s", err)
	}

	// create tx2 and add block
	tx2, err := bc.NewTransaction("me", "you", 20)
	if err != nil {
		cfg.Logger.Errorf("Failed to create UTXO transaction: %s", err)
	}
	_, err = bc.AddBlock([]*block.Transaction{tx2})
	if err != nil {
		cfg.Logger.Errorf("Failed to add block: %s", err)
	}

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
		cfg.Logger.Infof("PoW: %t", consensus.NewProofOfWork(1).ValidateBlock(blk))
	}

	cfg.Logger.Infof("Server listening on %s", cfg.BaseURL)
	err = http.ListenAndServe(cfg.BaseURL, NewHTTPRouter(bc))
	if err != nil {
		cfg.Logger.Fatalf("Failed to start server: %s", err)
	}
}
