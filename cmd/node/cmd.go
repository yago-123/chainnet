package main

import (
	"chainnet/config"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use: "chainnet-node",
	Run: func(cmd *cobra.Command, _ []string) {
		config.ApplyFlagsToConfig(cmd, cfg)
	},
}

func Execute(logger *logrus.Logger) {
	config.AddConfigFlags(rootCmd)

	if err := rootCmd.Execute(); err != nil {
		logger.Fatalf("error executing command: %v", err)
	}
}

func initConfig(logger *logrus.Logger) {
	var err error
	cfg, err = config.LoadConfig(cfgFile)
	if err != nil {
		logger.Infof("unable to find config file, relying in default config file...")
		cfg = config.NewConfig()
	}
}
