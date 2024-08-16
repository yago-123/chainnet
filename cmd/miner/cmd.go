package main

import (
	"chainnet/config"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use: "chainnet-miner",
	Run: func(cmd *cobra.Command, _ []string) {
		cfg = config.InitConfig(cmd)
	},
}

func Execute(logger *logrus.Logger) {
	config.AddConfigFlags(rootCmd)

	if err := rootCmd.Execute(); err != nil {
		logger.Fatalf("error executing command: %v", err)
	}
}