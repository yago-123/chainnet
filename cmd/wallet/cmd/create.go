package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

// createCmd represents the creation of wallets, addresses...
var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create wallet",
	Long:  `Create wallet`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Creating new wallet...")
	},
}

var createNewAddressCmd = &cobra.Command{
	Use:   "address",
	Short: "New address",
	Long:  `Create new address.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Creating address...")
	},
}

func init() {
	// main command
	rootCmd.AddCommand(createCmd)

	// sub commands
	createCmd.AddCommand(createNewAddressCmd)
}
