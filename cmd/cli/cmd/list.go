package cmd

import (
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/spf13/cobra"
)

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
	ep := fmt.Sprintf("%s/address/%s/transactions", BaseURL, url.PathEscape(address))
	response, err := http.Get(ep)
	if err != nil {
		logger.Infof("Error retrieving transactions endpoint: %s", err)
		return
	}
	defer response.Body.Close()
	body, _ := io.ReadAll(response.Body)
	logger.Infof("Transactions: %s", string(body))
}

func listUnspentTransactions(address string) {
	ep := fmt.Sprintf("%s/address/%s/utxos", BaseURL, url.PathEscape(address))
	response, err := http.Get(ep)
	if err != nil {
		logger.Infof("Error retrieving unspent transactions endpoint: %s", err)
		return
	}
	defer response.Body.Close()
	body, _ := io.ReadAll(response.Body)
	logger.Infof("Unspent transactions: %s", string(body))
}

func listBalance(address string) {
	ep := fmt.Sprintf("%s/address/%s/balance", BaseURL, url.PathEscape(address))
	response, err := http.Get(ep)
	if err != nil {
		logger.Infof("Error retrieving unspent transactions endpoint: %s", err)
		return
	}
	defer response.Body.Close()
	body, _ := io.ReadAll(response.Body)
	logger.Infof("Balance for %s: %s", address, string(body))
}
