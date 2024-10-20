package cmd

import (
	"context"

	"github.com/yago-123/chainnet/pkg/consensus/util"

	"github.com/spf13/cobra"
	"github.com/yago-123/chainnet/pkg/consensus/validator"
	"github.com/yago-123/chainnet/pkg/crypto/hash"
	"github.com/yago-123/chainnet/pkg/encoding"
	"github.com/yago-123/chainnet/pkg/kernel"
	wallt "github.com/yago-123/chainnet/pkg/wallet"
)

var err error

var sendCmd = &cobra.Command{
	Use:   "send",
	Short: "Send transaction",
	Long:  `Send transactions from wallets.`,
	Run: func(cmd *cobra.Command, _ []string) {
		address, _ := cmd.Flags().GetString("address")
		amount, _ := cmd.Flags().GetUint("amount")
		fee, _ := cmd.Flags().GetUint("fee")
		privKeyCont, _ := cmd.Flags().GetString("priv-key")
		privKeyPath, _ := cmd.Flags().GetString("priv-key-path")

		// check if only one private key is provided
		if (privKeyCont == "") != (privKeyPath == "") {
			logger.Fatalf("specify one argument containing the private key: --priv-key or --priv-key-path")
		}

		// process key from path or from content
		privKey := []byte{}
		if privKeyCont != "" {
			privKey = []byte(privKeyCont)
		}

		if privKeyPath != "" {
			privKey, err = util.ReadECDSAPemPrivateKey(cfg.P2P.Identity.PrivKeyPath)
			if err != nil {
				logger.Fatalf("error reading private key: %v", err)
			}
		}

		logger.Debugf(
			"Sending tx with address = %s, amount = %d, fee = %d, privKey = %x",
			address,
			amount,
			fee,
			privKey,
		)

		wallet, err := wallt.NewWallet(
			cfg,
			[]byte("1"),
			validator.NewLightValidator(hash.GetHasher(consensusHasherType)),
			consensusSigner,
			walletHasher,
			hash.GetHasher(consensusHasherType),
			encoding.NewProtobufEncoder(),
		)
		if err != nil {
			logger.Fatalf("error setting up wallet: %v", err)
		}

		tx, err := wallet.GenerateNewTransaction(address, amount, fee, []*kernel.UTXO{})
		if err != nil {
			logger.Fatalf("error generating transaction: %v", err)
		}

		err = wallet.SendTransaction(context.Background(), tx)
		if err != nil {
			logger.Fatalf("error sending transaction: %v", err)
		}
	},
}

func init() {
	// main command
	rootCmd.AddCommand(sendCmd)

	// sub commands
	sendCmd.Flags().String("address", "", "Destination address to send coins")
	sendCmd.Flags().Uint("amount", 0, "Amount of coins to send")
	sendCmd.Flags().Uint("fee", 0, "Amount of fee to send")
	sendCmd.Flags().String("priv-key", "", "Private key")
	sendCmd.Flags().String("priv-key-path", "", "Path to private key")

	// required flags
	_ = sendCmd.MarkFlagRequired("address")
	_ = sendCmd.MarkFlagRequired("amount")
}
