/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

// todo() use proper way with Cobra configurations
const BaseURL = "http://localhost:8080"

var rootCmd = &cobra.Command{
	Use:   "chainnet-cli",
	Short: "Chainnet CLI",
	Long:  `A CLI for interacting with chainnet.`,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

var address string

func init() {
	rootCmd.PersistentFlags().StringVarP(&address, "address", "a", "", "Address to use")
}
