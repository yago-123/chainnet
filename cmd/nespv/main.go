package main

import (
	"github.com/sirupsen/logrus"
	"github.com/yago-123/chainnet/cmd/nespv/cmd"
	"github.com/yago-123/chainnet/config"
	"github.com/yago-123/chainnet/pkg/consensus/validator"
	"github.com/yago-123/chainnet/pkg/crypto"
	"github.com/yago-123/chainnet/pkg/crypto/hash"
	"github.com/yago-123/chainnet/pkg/crypto/sign"
	"github.com/yago-123/chainnet/pkg/encoding"
	"github.com/yago-123/chainnet/pkg/wallet"
)

var cfg *config.Config

func main() {
	cmd.Execute(logrus.New())

	cfg.Logger.SetLevel(logrus.DebugLevel)

	cfg.Logger.Infof("starting wallet with config %v", cfg)

	// general consensus hasher (tx, block hashes...)
	consensusHasherType := hash.SHA256

	// general consensus signer (tx)
	consensusSigner := crypto.NewHashedSignature(
		sign.NewECDSASignature(),
		hash.NewSHA256(),
	)

	// algorithm required for wallet hashing
	walletHasher := crypto.NewMultiHash(
		[]hash.Hashing{hash.NewSHA256(), hash.NewRipemd160()},
	)

	// create new wallet for storing mining rewards
	_, err := wallet.NewWallet(
		cfg,
		[]byte("1"),
		validator.NewLightValidator(hash.GetHasher(consensusHasherType)),
		consensusSigner,
		walletHasher,
		hash.GetHasher(consensusHasherType),
		encoding.NewProtobufEncoder(),
	)
	if err != nil {
		cfg.Logger.Fatalf("Error creating new wallet: %s", err)
	}
}
