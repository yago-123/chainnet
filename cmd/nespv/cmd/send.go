package cmd

import (
	"context"

	"github.com/spf13/cobra"
	"github.com/yago-123/chainnet/pkg/consensus/validator"
	"github.com/yago-123/chainnet/pkg/crypto/hash"
	"github.com/yago-123/chainnet/pkg/encoding"
	"github.com/yago-123/chainnet/pkg/kernel"
	wallt "github.com/yago-123/chainnet/pkg/wallet"
)

var sendCmd = &cobra.Command{
	Use:   "send",
	Short: "Send transaction",
	Long:  `Send transactions from wallets.`,
	Run: func(cmd *cobra.Command, _ []string) {
		address, _ := cmd.Flags().GetString("address")
		amount, _ := cmd.Flags().GetUint("amount")
		fee, _ := cmd.Flags().GetUint("fee")
		privKeyPath, _ := cmd.Flags().GetString("priv-key-path")
		privKey, _ := cmd.Flags().GetString("priv-key")

		if privKey == "" && privKeyPath == "" {
			logger.Fatalf("specify at least one containing private key: --priv-key or --priv-key-path")
		}

		if privKey != "" {

		}

		if privKeyPath != "" {

		}

		logger.Infof(
			"Sending tx with priv.key = %s, priv. key path = %s, address = %s, amount = %d, fee = %d",
			privKey,
			privKeyPath,
			address,
			amount,
			fee,
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
	sendCmd.Flags().String("priv-key-path", "", "Path to private key")
	sendCmd.Flags().String("priv-key", "", "Private key")

	// required flags
	_ = sendCmd.MarkFlagRequired("address")
	_ = sendCmd.MarkFlagRequired("amount")
}
