package main

import (
	"chainnet/config"
	blockchain "chainnet/pkg/chain"
	"chainnet/pkg/chain/explorer"
	"chainnet/pkg/chain/observer"
	"chainnet/pkg/consensus/miner"
	"chainnet/pkg/consensus/validator"
	"chainnet/pkg/crypto"
	"chainnet/pkg/crypto/hash"
	"chainnet/pkg/crypto/sign"
	"chainnet/pkg/encoding"
	"chainnet/pkg/kernel"
	"chainnet/pkg/storage"
	"chainnet/pkg/wallet"
	"time"

	"github.com/sirupsen/logrus"
)

const MiningInterval = 15 * time.Second

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

	// create wallet address hasher
	walletSha256Ripemd160Hasher, err := crypto.NewMultiHash([]hash.Hashing{hash.NewSHA256(), hash.NewRipemd160()})
	if err != nil {
		cfg.Logger.Fatalf("Error creating multi-hash configuration: %s", err)
	}

	// create new wallet for storing mining rewards
	w, err := wallet.NewWallet(
		[]byte("1"),
		validator.NewLightValidator(hash.GetHasher(consensusHasherType)),
		crypto.NewHashedSignature(sign.NewECDSASignature(), hash.NewSHA256()),
		walletSha256Ripemd160Hasher,
		hash.GetHasher(consensusHasherType),
	)
	if err != nil {
		cfg.Logger.Fatalf("Error creating new wallet: %s", err)
	}

	// create instance for persisting data
	boltdb, err := storage.NewBoltDB("bin/miner-storage", "block-bucket", "header-bucket", encoding.NewGobEncoder())
	if err != nil {
		cfg.Logger.Fatalf("Error creating bolt db: %s", err)
	}

	// create new mempool
	mempool := miner.NewMemPool()

	// create new observer
	subjectObserver := observer.NewBlockSubject()

	// create new chain
	chain, err := blockchain.NewBlockchain(
		cfg,
		boltdb,
		hash.GetHasher(consensusHasherType),
		validator.NewHeavyValidator(
			validator.NewLightValidator(hash.GetHasher(consensusHasherType)),
			explorer.NewExplorer(boltdb),
			crypto.NewHashedSignature(
				sign.NewECDSASignature(), hash.NewSHA256(),
			),
			hash.GetHasher(consensusHasherType),
		),
		subjectObserver,
		encoding.NewProtobufEncoder(),
	)
	if err != nil {
		cfg.Logger.Fatalf("Error creating blockchain: %s", err)
	}

	// create new miner
	mine := miner.NewMiner(cfg, w.PublicKey, chain, mempool, hash.SHA256)

	// register chain observers
	subjectObserver.Register(mine)
	subjectObserver.Register(boltdb)
	subjectObserver.Register(mempool)

	for {
		time.Sleep(cfg.MiningInterval)

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
