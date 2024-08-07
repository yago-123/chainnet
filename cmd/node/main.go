package main

import (
	"chainnet/config"
	blockchain "chainnet/pkg/chain"
	"chainnet/pkg/chain/explorer"
	"chainnet/pkg/chain/observer"
	"chainnet/pkg/consensus/validator"
	"chainnet/pkg/crypto"
	"chainnet/pkg/crypto/hash"
	"chainnet/pkg/crypto/sign"
	"chainnet/pkg/encoding"
	"chainnet/pkg/storage"
	"time"

	"github.com/sirupsen/logrus"
)

const (
	MiningInterval = 1 * time.Minute
)

func main() {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	cfg := config.NewConfig(logger, MiningInterval, false, 0, 0)

	// general consensus hasher (tx, block hashes...)
	consensusHasherType := hash.SHA256

	// create instance for persisting data
	boltdb, err := storage.NewBoltDB("bin/miner-storage", "block-bucket", "header-bucket", encoding.NewGobEncoder())
	if err != nil {
		logger.Fatalf("Error creating bolt db: %s", err)
	}

	// create new observer
	subjectObserver := observer.NewSubjectObserver()

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
	)
	if err != nil {
		logger.Fatalf("Error creating blockchain: %s", err)
	}

	subjectObserver.Register(boltdb)

	chain.Sync() //?

	time.Sleep(MiningInterval)
}
