package cmd

import (
	"fmt"
	"net"

	"github.com/btcsuite/btcutil/base58"
	"github.com/spf13/cobra"
	sdkv1beta "github.com/yago-123/chainnet-sdk-go/v1beta"
	"github.com/yago-123/chainnet/config"
	sdkv1beta "github.com/yago-123/chainnet/pkg/sdk/v1beta"
)

const FlagAddress = "address"

var listTxsCmd = &cobra.Command{
	Use:   "list-txs",
	Short: "List all transactions",
	Run: func(cmd *cobra.Command, _ []string) {
		cfg = config.InitConfig(cmd)

		address, _ := cmd.Flags().GetString(FlagAddress)

		if len(address) == 0 {
			logger.Fatalf("address must be provided")
		}

		client, err := sdkv1beta.NewClient(
			net.JoinHostPort(cfg.Wallet.ServerAddress, fmt.Sprintf("%d", cfg.Wallet.ServerPort)),
			nil,
		)
		if err != nil {
			logger.Fatalf("failed to create SDK client: %v", err)
		}

		txs, err := client.GetAddressTransactions(cmd.Context(), base58.Decode(address))
		if err != nil {
			logger.Fatalf("failed to get transactions: %v", err)
		}

		// print transactions
		for _, tx := range txs {
			logger.Infof("{\n%+v}\n", tx)
		}
	},
}

func init() {
	// main command
	config.AddConfigFlags(listTxsCmd)
	rootCmd.AddCommand(listTxsCmd)

	// sub commands
	listTxsCmd.Flags().String(FlagAddress, "", "Destination address to send coins")

	// required flags
	_ = listTxsCmd.MarkFlagRequired(FlagAddress)
}
