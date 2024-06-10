package cli

import (
	"fmt"
	"github.com/spf13/cobra"
	"io/ioutil"
	"net/http"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List information",
	Run: func(cmd *cobra.Command, args []string) {
		if txs {
			listTransactions()
		} else if utxos {
			listUTXOs()
		} else if blocks {
			listBlocks()
		} else {
			fmt.Println("Please specify a valid flag: --txs, --utxos, --blocks")
		}
	},
}

var txs, utxos, blocks bool

func init() {
	rootCmd.AddCommand(listCmd)
	listCmd.Flags().BoolVarP(&txs, "txs", "t", false, "List transactions")
	listCmd.Flags().BoolVarP(&utxos, "utxos", "u", false, "List UTXOs")
	listCmd.Flags().BoolVarP(&blocks, "blocks", "b", false, "List blocks")
}

func listTransactions() {
	url := fmt.Sprintf("%s/transactions", cfg.BaseURL)
	response, err := http.Get(url)
	if err != nil {
		cfg.Logger.Error("Error:", err)
		return
	}
	defer response.Body.Close()
	body, _ := ioutil.ReadAll(response.Body)
	fmt.Println("Transactions:", string(body))
}

func listUTXOs() {
	url := fmt.Sprintf("%s/utxos", cfg.BaseURL)
	response, err := http.Get(url)
	if err != nil {
		cfg.Logger.Error("Error:", err)
		return
	}
	defer response.Body.Close()
	body, _ := ioutil.ReadAll(response.Body)
	fmt.Println("UTXOs:", string(body))
}

func listBlocks() {
	url := fmt.Sprintf("%s/blocks", cfg.BaseURL)
	response, err := http.Get(url)
	if err != nil {
		cfg.Logger.Error("Error:", err)
		return
	}
	defer response.Body.Close()
	body, _ := ioutil.ReadAll(response.Body)
	fmt.Println("Blocks:", string(body))
}
