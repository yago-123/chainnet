package main

import (
	"chainnet/config"
	blockchain "chainnet/pkg/chain"
	"chainnet/pkg/chain/explorer"
	"chainnet/pkg/consensus/validator"
	"chainnet/pkg/crypto"
	"chainnet/pkg/crypto/hash"
	"chainnet/pkg/crypto/sign"
	"chainnet/pkg/encoding"
	"chainnet/pkg/kernel"
	"chainnet/pkg/mempool"
	"chainnet/pkg/miner"
	"chainnet/pkg/observer"
	"chainnet/pkg/storage"
	"time"

	"github.com/sirupsen/logrus"
)

var cfg *config.Config

func main() {
	var block *kernel.Block

	// execute the root command
	Execute(logrus.New())

	cfg.Logger.SetLevel(logrus.DebugLevel)

	cfg.Logger.Infof("starting chain node with config %v", cfg)

	// general consensus hasher (tx, block hashes...)
	consensusHasherType := hash.SHA256
	// todo(): add consensusSignatureType

	// create observer controllers
	subjectChain := observer.NewChainSubject()
	subjectNet := observer.NewNetSubject()

	// create instance for persisting data
	boltdb, err := storage.NewBoltDB(cfg.StorageFile, "block-bucket", "header-bucket", encoding.NewGobEncoder())
	if err != nil {
		cfg.Logger.Fatalf("Error creating bolt db: %s", err)
	}

	// create mempool instance
	mempool := mempool.NewMemPool()

	// create new chain
	chain, err := blockchain.NewBlockchain(
		cfg,
		boltdb,
		mempool,
		hash.GetHasher(consensusHasherType),
		validator.NewHeavyValidator(
			validator.NewLightValidator(hash.GetHasher(consensusHasherType)),
			explorer.NewExplorer(boltdb, hash.GetHasher(consensusHasherType)),
			crypto.NewHashedSignature(
				sign.NewECDSASignature(), hash.NewSHA256(),
			),
			hash.GetHasher(consensusHasherType),
		),
		subjectChain,
		encoding.NewProtobufEncoder(),
	)
	if err != nil {
		cfg.Logger.Fatalf("Error creating blockchain: %s", err)
	}

	// create new miner
	mine, err := miner.NewMiner(cfg, chain, hash.SHA256)
	if err != nil {
		cfg.Logger.Fatalf("error initializing miner: %s", err)
	}

	// register network observers
	subjectNet.Register(chain)

	// register chain observers
	subjectChain.Register(mine)
	subjectChain.Register(boltdb)
	subjectChain.Register(mempool)

	network, err := chain.InitNetwork(subjectNet)
	if err != nil {
		cfg.Logger.Fatalf("Error initializing network: %s", err)
	}

	// register the block subject to the network
	subjectChain.Register(network)

	for {
		// start mining block
		block, err = mine.MineBlock()
		if err != nil {
			cfg.Logger.Errorf("Error mining block: %s", err)
			continue
		}

		miningTime := time.Unix(block.Header.Timestamp, 0).Format(time.RFC3339)

		if block.IsGenesisBlock() {
			cfg.Logger.Infof(
				"genesis block mined successfully: hash %x, number txs %d, time %s, height %d, target %d, nonce %d",
				block.Hash, len(block.Transactions), miningTime, block.Header.Height, block.Header.Target, block.Header.Nonce,
			)
			continue
		}

		cfg.Logger.Infof(
			"block mined successfully: hash %x, previous hash %x, number txs %d, time %s, height %d, target %d, nonce %d",
			block.Hash, block.Header.PrevBlockHash, len(block.Transactions), miningTime, block.Header.Height, block.Header.Target, block.Header.Nonce,
		)
	}
}
