package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"io/ioutil"
	"net/http"
)

var wallet string

var balanceCmd = &cobra.Command{
	Use:   "balance",
	Short: "Retrieve the balance for a given wallet",
	Run: func(cmd *cobra.Command, args []string) {
		url := fmt.Sprintf("%s/wallet/%s/balance", cfg.BaseURL, wallet)
		response, err := http.Get(url)
		if err != nil {
			cfg.Logger.Error("Error:", err)
			return
		}
		defer response.Body.Close()
		body, _ := ioutil.ReadAll(response.Body)
		fmt.Println("Balance:", string(body))
	},
}

func init() {
	rootCmd.AddCommand(balanceCmd)
	balanceCmd.Flags().StringVarP(&wallet, "wallet", "w", "", "Wallet address to retrieve balance for")
	balanceCmd.MarkFlagRequired("wallet")
}
