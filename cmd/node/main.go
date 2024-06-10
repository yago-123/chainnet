package node

import (
	"chainnet/config"
	block "chainnet/pkg/block"
	"chainnet/pkg/chain"
	"chainnet/pkg/consensus"
	"chainnet/pkg/encoding"
	"chainnet/pkg/storage"
	"fmt"
	"github.com/sirupsen/logrus"
	"log"
	"math"
	"net/http"
)

const (
	MINING_DIFFICULTY = 1
	MAX_NONCE         = math.MaxInt64
)

func main() {
	logger := logrus.New()
	cfg := config.NewConfig(logger, MINING_DIFFICULTY, MAX_NONCE, "http://localhost:8080")

	bc := blockchain.NewBlockchain(cfg, consensus.NewProofOfWork(cfg), storage.NewBoltDB("_fixture/chainnet-store", "chainnet-bucket", encoding.NewGobEncoder(logger), logger))

	bc.AddBlock([]*block.Transaction{block.NewCoinbaseTx("me", "data")})
	bc.AddBlock([]*block.Transaction{})
	bc.AddBlock([]*block.Transaction{})

	iterator := bc.CreateIterator()
	for iterator.HasNext() {
		block, err := iterator.GetNextBlock()
		if err != nil {
			logger.Panicf("Error getting block: %s", err)
		}

		logger.Infof("----------------------")
		logger.Infof("Prev. hash: %x", block.PrevBlockHash)
		logger.Infof("Num transactions: %d", len(block.Transactions))
		logger.Infof("Hash: %x", block.Hash)
		logger.Infof("PoW: %t", consensus.NewProofOfWork(cfg).Validate(block))
	}

	router := NewRouter()

	fmt.Println("Server listening on %s", cfg.BaseURL)
	log.Fatal(http.ListenAndServe(cfg.BaseURL, router))
}
