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
	"chainnet/pkg/mempool"
	"chainnet/pkg/observer"
	"chainnet/pkg/storage"

	"github.com/sirupsen/logrus"
)

var cfg *config.Config

func main() {
	var err error

	// execute the root command
	Execute(logrus.New())

	cfg.Logger.SetLevel(logrus.DebugLevel)

	cfg.Logger.Infof("starting chain node with config %v", cfg)

	// general consensus hasher (tx, block hashes...)
	consensusHasherType := hash.SHA256

	// create new observer
	netSubject := observer.NewNetSubject()
	subjectChain := observer.NewBlockSubject()

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
			validator.NewLightValidator(hash.GetHasher(consensusHasherType)),
			explorer.NewExplorer(boltdb),
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
