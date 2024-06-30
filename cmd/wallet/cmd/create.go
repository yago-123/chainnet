package cmd

import (
	"chainnet/pkg/consensus"
	"chainnet/pkg/crypto"
	"chainnet/pkg/crypto/hash"
	"chainnet/pkg/crypto/sign"
	"chainnet/pkg/wallet"

	"github.com/spf13/cobra"
)

// createCmd represents the creation of wallets, addresses...
var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create wallet",
	Long:  `Create wallet`,
	Run: func(_ *cobra.Command, _ []string) {
		logger.Infof("Creating new wallet...")

		sha256Ripemd160Hasher, err := crypto.NewMultiHash([]hash.Hashing{hash.NewSHA256(), hash.NewRipemd160()})
		if err != nil {
			logger.Infof("Error creating new wallet: %s", err)
		}

		ecdsaSha256Signer := crypto.NewHashedSignature(sign.NewECDSASignature(), hash.NewSHA256())

		w, err := wallet.NewWallet(
			[]byte("0.0.1"),
			consensus.NewLightValidator(),
			ecdsaSha256Signer,
			sha256Ripemd160Hasher,
		)

		if err != nil {
			logger.Infof("Error creating new wallet: %s", err)
		}

		logger.Infof("Created wallet %s", w.ID())
	},
}

var createNewAddressCmd = &cobra.Command{
	Use:   "address",
	Short: "New address",
	Long:  `Create new address.`,
	Run: func(_ *cobra.Command, _ []string) {
		logger.Infof("Creating address...")
	},
}

func init() {
	// main command
	rootCmd.AddCommand(createCmd)

	// sub commands
	createCmd.AddCommand(createNewAddressCmd)
}
