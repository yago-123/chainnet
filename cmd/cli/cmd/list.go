package cmd

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/spf13/cobra"
)

const RequestTimeout = 5 * time.Second

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List information",
	Long:  `List information`,
	Run: func(_ *cobra.Command, _ []string) {

	},
}

var listTxsCmd = &cobra.Command{
	Use:   "txs",
	Short: "Transactions",
	Long:  `List transactions.`,
	Run: func(_ *cobra.Command, _ []string) {
		// todo() list all transfers if address == ""
		if address == "" {
			logger.Infof("can't retrieve transactions, use --address flag")
		}

		if address != "" {
			listTransactions(address)
		}
	},
}

var listUTXOsCmd = &cobra.Command{
	Use:   "utxos",
	Short: "Unspent transactions",
	Long:  "List unspent transactions.",
	Run: func(_ *cobra.Command, _ []string) {
		// todo() list all utxos if address == ""
		if address == "" {
			logger.Infof("can't retrieve unspent transactions, use --address flag")
		}

		if address != "" {
			listUnspentTransactions(address)
		}
	},
}

var listBalanceCmd = &cobra.Command{
	Use:   "balance",
	Short: "Balance",
	Long:  "List balance.",
	Run: func(_ *cobra.Command, _ []string) {
		if address == "" {
			logger.Infof("can't list balance, specifiy with --address flag")
		}

		if address != "" {
			listBalance(address)
		}
	},
}

func init() {
	// main command
	rootCmd.AddCommand(listCmd)

	// sub commands
	listCmd.AddCommand(listTxsCmd)
	listCmd.AddCommand(listUTXOsCmd)
	listCmd.AddCommand(listBalanceCmd)
}

func listTransactions(address string) {
	ctx, cancel := context.WithTimeout(context.Background(), RequestTimeout)
	defer cancel()

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, generateEndpoint(BaseURL, address, "transactions"), nil)
	if err != nil {
		logger.Infof("Error retrieving transactions endpoint: %s", err)
		return
	}
	defer request.Body.Close()
	body, _ := io.ReadAll(request.Body)
	logger.Infof("Transactions: %s", string(body))
}

func listUnspentTransactions(address string) {
	ctx, cancel := context.WithTimeout(context.Background(), RequestTimeout)
	defer cancel()

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, generateEndpoint(BaseURL, address, "utxos"), nil)
	if err != nil {
		logger.Infof("Error retrieving unspent transactions endpoint: %s", err)
		return
	}
	defer request.Body.Close()
	body, _ := io.ReadAll(request.Body)
	logger.Infof("Unspent transactions: %s", string(body))
}

func listBalance(address string) {
	ctx, cancel := context.WithTimeout(context.Background(), RequestTimeout)
	defer cancel()

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, generateEndpoint(BaseURL, address, "balance"), nil)
	if err != nil {
		logger.Infof("Error retrieving unspent transactions endpoint: %s", err)
		return
	}
	defer request.Body.Close()
	body, _ := io.ReadAll(request.Body)
	logger.Infof("Balance for %s: %s", address, string(body))
}

func generateEndpoint(baseURL, address, target string) string {
	return fmt.Sprintf("%s/address/%s/%s", baseURL, url.PathEscape(address), target)
}
