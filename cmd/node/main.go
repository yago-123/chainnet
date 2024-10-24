package main

import (
	"github.com/sirupsen/logrus"
	"github.com/yago-123/chainnet/config"
	blockchain "github.com/yago-123/chainnet/pkg/chain"
	"github.com/yago-123/chainnet/pkg/chain/explorer"
	"github.com/yago-123/chainnet/pkg/consensus/validator"
	"github.com/yago-123/chainnet/pkg/crypto"
	"github.com/yago-123/chainnet/pkg/crypto/hash"
	"github.com/yago-123/chainnet/pkg/crypto/sign"
	"github.com/yago-123/chainnet/pkg/encoding"
	"github.com/yago-123/chainnet/pkg/mempool"
	"github.com/yago-123/chainnet/pkg/observer"
	"github.com/yago-123/chainnet/pkg/storage"
)

var cfg *config.Config

var (
	// general consensus hasher (tx, block hashes...)
	consensusHasherType = hash.SHA256

	// general consensus signer (tx)
	consensusSigner = crypto.NewHashedSignature(
		sign.NewECDSASignature(),
		hash.NewSHA256(),
	)
)

func main() {
	var err error

	// execute the root command
	Execute(logrus.New())

	cfg.Logger.SetLevel(logrus.DebugLevel)

	cfg.Logger.Infof("starting chain node with config %v", cfg)

	// create new observer
	netSubject := observer.NewNetSubject()
	subjectChain := observer.NewChainSubject()

	// create instance for persisting data
	boltdb, err := storage.NewBoltDB(cfg.StorageFile, "block-bucket", "header-bucket", encoding.NewGobEncoder())
	if err != nil {
		cfg.Logger.Fatalf("Error creating bolt db: %s", err)
	}

	subjectChain.Register(boltdb)

	// create new chain
	chain, err := blockchain.NewBlockchain(
		cfg,
		boltdb,
		mempool.NewMemPool(),
		hash.GetHasher(consensusHasherType),
		validator.NewHeavyValidator(
			cfg,
			validator.NewLightValidator(hash.GetHasher(consensusHasherType)),
			explorer.NewExplorer(boltdb, hash.GetHasher(consensusHasherType)),
			consensusSigner,
			hash.GetHasher(consensusHasherType),
		),
		subjectChain,
		encoding.NewProtobufEncoder(),
	)
	if err != nil {
		cfg.Logger.Fatalf("Error creating blockchain: %s", err)
	}

	// create net subject and register chain
	netSubject.Register(chain)

	network, err := chain.InitNetwork(netSubject)
	if err != nil {
		cfg.Logger.Fatalf("Error initializing network: %s", err)
	}

	// register the block subject to the network
	subjectChain.Register(network)

	select {}
}
