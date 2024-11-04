package cmd

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/yago-123/chainnet/config"
)

var cfg *config.Config
var logger = logrus.New()

var rootCmd = &cobra.Command{
	Use: "chainnet-cli",
	Run: func(_ *cobra.Command, _ []string) {

	},
}

func Execute(logger *logrus.Logger) {
	if err := rootCmd.Execute(); err != nil {
		logger.Fatalf("error executing command: %v", err)
	}
}
