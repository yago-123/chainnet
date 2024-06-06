package main

import (
	"chainnet/chain"
	"chainnet/config"
	"chainnet/consensus"
	"chainnet/encoding"
	"chainnet/storage"
	"github.com/sirupsen/logrus"
	"math"
)

const (
	MINING_DIFFICULTY = 1
	MAX_NONCE         = math.MaxInt64
)

func main() {
	logger := logrus.New()
	cfg := config.NewConfig(logger, MINING_DIFFICULTY, MAX_NONCE)

	bc := blockchain.NewBlockchain(cfg, consensus.NewProofOfWork(cfg), storage.NewBoltDB("_fixture/chainnet-store", "chainnet-bucket", encoding.NewGobEncoder(logger), logger))

	bc.AddBlock("Send 1 BTC to Ivan")
	bc.AddBlock("Send 2 more BTC to Ivan")

	for _, blockHash := range bc.Chain {
		logger.Infof("----------------------")
		logger.Infof("Prev. hash: %x", bc.Blocks[blockHash].PrevBlockHash)
		logger.Infof("Data: %s", bc.Blocks[blockHash].Data)
		logger.Infof("Hash: %x", bc.Blocks[blockHash].Hash)
		logger.Infof("PoW: %t", consensus.NewProofOfWork(cfg).Validate(bc.Blocks[blockHash]))
	}
}
