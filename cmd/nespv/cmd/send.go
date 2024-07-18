package cmd

import (
	"github.com/spf13/cobra"
)

var sendCmd = &cobra.Command{
	Use:   "send",
	Short: "Send transaction",
	Long:  `Send transactions from wallets.`,
	Run: func(_ *cobra.Command, _ []string) {
		logger.Infof("Sending transactions...")
	},
}

func init() {
	// main command
	rootCmd.AddCommand(sendCmd)
}
