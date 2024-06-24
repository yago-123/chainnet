package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List wallets",
	Long:  `List wallets`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Listing wallets...")
	},
}

var listAddressesCmd = &cobra.Command{
	Use:   "address",
	Short: "List addresses",
	Long:  `Lising addresses of wallet.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Listing addresses...")
	},
}

var listBalanceCmd = &cobra.Command{
	Use:   "balance",
	Short: "List balance",
	Long:  `List balance of wallet.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Listing balance...")
	},
}

func init() {
	// main command
	rootCmd.AddCommand(listCmd)

	// sub commands
	createCmd.AddCommand(listAddressesCmd)
	createCmd.AddCommand(listBalanceCmd)
}
