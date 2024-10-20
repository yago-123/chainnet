package main

import (
	"github.com/sirupsen/logrus"
	"github.com/yago-123/chainnet/cmd/nespv/cmd"
)

func main() {
	cmd.Execute(logrus.New())
	/*
		cmd.Execute(logrus.New())

		// todo() remove this cfg initialization
		cfg = config.NewConfig()

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
	*/
}
