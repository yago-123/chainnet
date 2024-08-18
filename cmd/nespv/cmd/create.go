package cmd

import (
	"chainnet/config"
	"chainnet/pkg/consensus/validator"
	"chainnet/pkg/crypto"
	"chainnet/pkg/crypto/hash"
	"chainnet/pkg/crypto/sign"
	"chainnet/pkg/encoding"
	"chainnet/pkg/wallet"

	"github.com/spf13/cobra"
)

// createCmd represents the creation of wallets, addresses...
var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create nespv",
	Long:  `Create nespv`,
	Run: func(_ *cobra.Command, _ []string) {
		logger.Infof("Creating new nespv...")

		sha256Ripemd160Hasher, err := crypto.NewMultiHash([]hash.Hashing{hash.NewSHA256(), hash.NewRipemd160()})
		if err != nil {
			logger.Infof("Error creating new nespv: %s", err)
		}

		ecdsaSha256Signer := crypto.NewHashedSignature(sign.NewECDSASignature(), hash.NewSHA256())

		w, err := wallet.NewWallet(
			config.NewConfig(), // todo() change this to a real config
			[]byte("0.0.1"),
			validator.NewLightValidator(hash.NewSHA256()),
			ecdsaSha256Signer,
			sha256Ripemd160Hasher,
			hash.NewSHA256(),
			encoding.NewProtobufEncoder(),
		)

		if err != nil {
			logger.Infof("Error creating new nespv: %s", err)
		}

		logger.Infof("Created nespv %s", w.ID())
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
