package main

import (
	"chainnet/config"
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

var rootCmd = &cobra.Command{
	Use:   "node",
	Short: "Chainnet node app",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func initConfig() {
	var err error
	cfg, err = config.LoadConfig(cfgFile)
	if err == nil {
		// success reading config file in file system
		return
	}

	fmt.Printf("unable to find config file: %s\n", err)

	fmt.Println("relying in default config file...")
	cfg = config.NewConfig()
}
