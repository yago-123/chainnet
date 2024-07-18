package cmd

import (
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// todo() use proper way with Cobra configurations
const BaseURL = "http://localhost:8080"

var logger *logrus.Logger

var rootCmd = &cobra.Command{
	Use:   "chainnet-nespv",
	Short: "Chainnet Wallet",
	Long:  `A nespv for interacting with chainnet.`,
}

func Execute(loggerHandler *logrus.Logger) {
	if loggerHandler == nil {
		return
	}

	logger = loggerHandler
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

var endpoint string
var profile string

func init() {
	rootCmd.PersistentFlags().StringVarP(&endpoint, "endpoint", "e", BaseURL, "Default endpoint to connect")
	rootCmd.PersistentFlags().StringVarP(&profile, "profile", "p", "default", "Wallet profile to use")
}
