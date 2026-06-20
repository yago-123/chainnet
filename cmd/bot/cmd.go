package main

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/yago-123/chainnet/config"
	botconfig "github.com/yago-123/chainnet/config/bot"
)

const (
	legacyBotKeyPathFlag      = "key-path"
	legacyBotMetadataPathFlag = "metadata"
)

var rootCmd = &cobra.Command{
	Use: "chainnet-bot",
	Run: func(cmd *cobra.Command, _ []string) {
		botCfg = botconfig.InitConfig(cmd)
		cfg = newChainConfigFromBotConfig(botCfg)
		applyLegacyBotFlags(cmd)
	},
}

func Execute(logger *logrus.Logger) {
	botconfig.AddConfigFlags(rootCmd)
	addLegacyBotFlags(rootCmd)

	if err := rootCmd.Execute(); err != nil {
		logger.Fatalf("error executing command: %v", err)
	}
}

func addLegacyBotFlags(cmd *cobra.Command) {
	cmd.Flags().String(legacyBotKeyPathFlag, botconfig.DefaultKeyPath, "Path to the HD wallet key")
	cmd.Flags().String(legacyBotMetadataPathFlag, botconfig.DefaultMetadataPath, "Path to the HD wallet metadata")
}

func applyLegacyBotFlags(cmd *cobra.Command) {
	if cmd.Flags().Changed(legacyBotKeyPathFlag) {
		keyPath, err := cmd.Flags().GetString(legacyBotKeyPathFlag)
		if err != nil {
			cfg.Logger.Fatalf("error reading %s flag: %v", legacyBotKeyPathFlag, err)
		}

		botCfg.Bot.KeyPath = keyPath
	}

	if cmd.Flags().Changed(legacyBotMetadataPathFlag) {
		metadataPath, err := cmd.Flags().GetString(legacyBotMetadataPathFlag)
		if err != nil {
			cfg.Logger.Fatalf("error reading %s flag: %v", legacyBotMetadataPathFlag, err)
		}

		botCfg.Bot.MetadataPath = metadataPath
	}
}

func newChainConfigFromBotConfig(botCfg *botconfig.Config) *config.Config {
	cfg := config.NewConfig()
	cfg.Logger = botCfg.Logger
	cfg.Wallet = botCfg.Wallet

	return cfg
}
