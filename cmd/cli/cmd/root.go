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
	Use:   "chainnet-cli",
	Short: "Chainnet CLI",
	Long:  `A CLI for interacting with chainnet.`,
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

var address string

func init() {
	rootCmd.PersistentFlags().StringVarP(&address, "address", "a", "", "Address to use")
}
