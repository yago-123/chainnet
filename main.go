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

	bc.AddBlock("Create Genesis block")
	bc.AddBlock("Send 1 BTC to Ivan")
	bc.AddBlock("Send 2 more BTC to Ivan")

	for _, blockHash := range bc.Chain {
		block, err := bc.GetBlock(blockHash)
		if err != nil {
			logger.Panicf("Error getting block: %s", err)
		}
		logger.Infof("----------------------")
		logger.Infof("Prev. hash: %x", block.PrevBlockHash)
		logger.Infof("Data: %s", block.Data)
		logger.Infof("Hash: %x", block.Hash)
		logger.Infof("PoW: %t", consensus.NewProofOfWork(cfg).Validate(block))
	}
}
