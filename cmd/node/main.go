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
	"chainnet/pkg/wallet"
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
	// todo(): add consensusSignatureType

	// create wallet address hasher
	walletSha256Ripemd160Hasher, err := crypto.NewMultiHash([]hash.Hashing{hash.NewSHA256(), hash.NewRipemd160()})
	if err != nil {
		logger.Fatalf("Error creating multi-hash configuration: %s", err)
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
		logger.Fatalf("Error creating new wallet: %s", err)
	}

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
