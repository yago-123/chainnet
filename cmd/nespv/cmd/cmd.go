package cmd

import (
	"crypto/sha256"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/yago-123/chainnet/config"
	"github.com/yago-123/chainnet/pkg/crypto"
	"github.com/yago-123/chainnet/pkg/crypto/hash"
	"github.com/yago-123/chainnet/pkg/crypto/sign"
)

var cfg *config.Config
var logger = logrus.New()

var (
	// general consensus hasher (tx, block hashes...)
	consensusHasherType = hash.SHA256

	// general consensus signer (tx)
	consensusSigner = crypto.NewHashedSignature(
		sign.NewECDSASignature(),
		hash.NewHasher(sha256.New()),
	)
)

var rootCmd = &cobra.Command{
	Use: "chainnet-nespv",
	Run: func(_ *cobra.Command, _ []string) {

	},
}

func Execute(logger *logrus.Logger) {
	if err := rootCmd.Execute(); err != nil {
		logger.Fatalf("error executing command: %v", err)
	}
}
