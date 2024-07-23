package main

import (
	"chainnet/config"
	"chainnet/pkg/chain/explorer"
	"chainnet/pkg/encoding"
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
	bolt, err := storage.NewBoltDB("_fixture/chainnet-store", "chainnet-bucket", encoding.NewGobEncoder())
	if err != nil {
		cfg.Logger.Errorf("Failed to initialize BoltDB: %s", err)
	}

	// create blockchain
	// bc := blockchain.NewBlockchain(cfg, miner.NewProofOfWork(cfg.DifficultyPoW, hash.NewSHA256()), bolt, &mockConsensus.MockHeavyValidator{})

	cfg.Logger.Infof("Server listening on %s", cfg.BaseURL)

	explorer := explorer.NewExplorer(bolt)
	err = http.ListenAndServe(cfg.BaseURL, NewHTTPRouter(explorer)) //nolint:gosec // add timeout later
	if err != nil {
		cfg.Logger.Fatalf("Failed to start server: %s", err)
	}
}
