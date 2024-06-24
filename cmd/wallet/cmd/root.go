package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

// todo() use proper way with Cobra configurations
const BaseURL = "http://localhost:8080"

var rootCmd = &cobra.Command{
	Use:   "chainnet-wallet",
	Short: "Chainnet Wallet",
	Long:  `A wallet for interacting with chainnet.`,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

var endpoint string
var profile string

func init() {
	rootCmd.PersistentFlags().StringVarP(&endpoint, "endpoint", "ep", BaseURL, "Default endpoint to connect")
	rootCmd.PersistentFlags().StringVarP(&profile, "profile", "p", "", "Wallet profile to use")
}
