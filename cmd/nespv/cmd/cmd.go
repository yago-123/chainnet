package cmd

import (
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
	consensusHasherType = hash.SHA256
	consensusSigner     = crypto.NewHashedSignature(
		sign.NewECDSASignature(),
		hash.NewSHA256(),
	)
	walletHasher = crypto.NewMultiHash(
		[]hash.Hashing{hash.NewSHA256(), hash.NewRipemd160()},
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
