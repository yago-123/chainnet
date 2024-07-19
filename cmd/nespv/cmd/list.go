package cmd

import (
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List wallets",
	Long:  `List wallets`,
	Run: func(_ *cobra.Command, _ []string) {
		logger.Infof("Listing wallets...")
	},
}

var listAddressesCmd = &cobra.Command{
	Use:   "address",
	Short: "List addresses",
	Long:  `Lising addresses of nespv.`,
	Run: func(_ *cobra.Command, _ []string) {
		logger.Infof("Listing addresses...")
	},
}

var listBalanceCmd = &cobra.Command{
	Use:   "balance",
	Short: "List balance",
	Long:  `List balance of nespv.`,
	Run: func(_ *cobra.Command, _ []string) {
		logger.Infof("Listing balance...")
	},
}

func init() {
	// main command
	rootCmd.AddCommand(listCmd)

	// sub commands
	createCmd.AddCommand(listAddressesCmd)
	createCmd.AddCommand(listBalanceCmd)
}
