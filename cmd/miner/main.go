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
	"chainnet/pkg/storage"
	"chainnet/pkg/wallet"
	"github.com/sirupsen/logrus"
)

func main() {
	logger := logrus.New()

	sha256Ripemd160Hasher, err := crypto.NewMultiHash([]hash.Hashing{hash.NewSHA256(), hash.NewRipemd160()})
	if err != nil {
		logger.Errorf("Error creating multi-hash configuration: %s", err)
	}

	w, err := wallet.NewWallet(
		[]byte("0.0.1"),
		validator.NewLightValidator(hash.NewSHA256()),
		crypto.NewHashedSignature(sign.NewECDSASignature(), hash.NewSHA256()),
		sha256Ripemd160Hasher,
	)
	if err != nil {
		logger.Errorf("Error creating new wallet: %s", err)
	}

	boltdb, err := storage.NewBoltDB("boltdb-file", "block-bucket", "header-bucket", encoding.NewGobEncoder())
	chain, err := blockchain.NewBlockchain(
		&config.Config{},
		boltdb,
		validator.NewHeavyValidator(validator.NewLightValidator(hash.NewSHA256()), *explorer.NewExplorer(boltdb), crypto.NewHashedSignature(sign.NewECDSASignature(), hash.NewSHA256()), sha256Ripemd160Hasher),
		observer.NewSubjectObserver(),
	)
	if err != nil {
		logger.Errorf("Error creating blockchain: %s", err)
	}

	mine := miner.NewMiner(w.PublicKey, hash.SHA256, chain)

}
