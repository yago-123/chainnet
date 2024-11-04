package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/yago-123/chainnet/config"
	"github.com/yago-123/chainnet/pkg/encoding"
	"github.com/yago-123/chainnet/pkg/net"
	"io"
	"net/http"
)

const FlagAddress = "address"

var listTxsCmd = &cobra.Command{
	Use:   "list-txs",
	Short: "List all transactions",
	Run: func(cmd *cobra.Command, args []string) {
		cfg = config.InitConfig(cmd)

		address, _ := cmd.Flags().GetString(FlagAddress)

		if len(address) == 0 {
			logger.Fatalf("address is required")
		}

		url := fmt.Sprintf(
			"http://%s:%d/%s",
			cfg.Wallet.ServerAddress,
			cfg.Wallet.ServerPort,
			fmt.Sprintf(net.RouterRetrieveAddressTxs, address),
		)

		// send request
		resp, err := http.Get(url)
		if err != nil {
			logger.Fatalf("failed to get transactions: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			logger.Fatalf("failed to get transactions: %v", resp.Status)
		}

		// decode response
		encoder := encoding.NewProtobufEncoder()
		data, err := io.ReadAll(resp.Body)
		if err != nil {
			logger.Fatalf("failed to read response body: %v", err)
		}

		txs, err := encoder.DeserializeTransactions(data)
		if err != nil {
			logger.Fatalf("failed to deserialize transactions: %v", err)
		}

		// print transactions
		for _, tx := range txs {
			logger.Infof("%s", tx.String())
		}
	},
}

func init() {
	// main command
	config.AddConfigFlags(listTxsCmd)
	rootCmd.AddCommand(listTxsCmd)

	// sub commands
	// todo() change address to pub key?
	listTxsCmd.Flags().String(FlagAddress, "", "Destination address to send coins")

	// required flags
	_ = listTxsCmd.MarkFlagRequired(FlagAddress)
}
