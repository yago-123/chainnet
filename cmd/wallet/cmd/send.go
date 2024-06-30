package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var sendCmd = &cobra.Command{
	Use:   "send",
	Short: "Send transaction",
	Long:  `Send transactions from wallets.`,
	Run: func(_ *cobra.Command, _ []string) {
		fmt.Println("Sending transactions...")
	},
}

func init() {
	// main command
	rootCmd.AddCommand(sendCmd)
}
