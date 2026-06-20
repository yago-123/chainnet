package bot

import (
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	walletcommon "github.com/yago-123/chainnet/pkg/wallet"
)

const (
	KeyConfigFile = "config"

	KeyWalletServerAddress  = "wallet.server-address"
	KeyWalletServerPort     = "wallet.server-port"
	KeyWalletRequestTimeout = "wallet.request-timeout"

	KeyBotKeyPath                      = "bot.key-path"
	KeyBotMetadataPath                 = "bot.metadata-path"
	KeyBotMaxConcurrentAccounts        = "bot.max-concurrent-accounts"
	KeyBotMaxWalletsPerAccount         = "bot.max-wallets-per-account"
	KeyBotMinimumTxBalance             = "bot.minimum-tx-balance"
	KeyBotMinStartupFundDistribution   = "bot.min-startup-fund-distribution"
	KeyBotMaxStartupFundDistribution   = "bot.max-startup-fund-distribution"
	KeyBotMinTimeBetweenTransactions   = "bot.min-time-between-transactions"
	KeyBotMaxTimeBetweenTransactions   = "bot.max-time-between-transactions"
	KeyBotSendTransactionTimeout       = "bot.send-transaction-timeout"
	KeyBotMetadataBackupPeriod         = "bot.metadata-backup-period"
	KeyBotMaxInputGroupsForCreatingTx  = "bot.max-input-groups-for-creating-tx"
	KeyBotMaxOutputGroupsForCreatingTx = "bot.max-output-groups-for-creating-tx"
)

const (
	DefaultConfigFile = ""

	DefaultKeyPath                      = "wallet.pem"
	DefaultMetadataPath                 = "hd_wallet.data"
	DefaultMaxConcurrentAccounts        = uint(15)
	DefaultMaxWalletsPerAccount         = uint(5)
	DefaultMinimumTxBalance             = uint(100000000)
	DefaultMinStartupFundDistribution   = 1 * time.Second
	DefaultMaxStartupFundDistribution   = 500 * time.Second
	DefaultMinTimeBetweenTransactions   = 60 * time.Second
	DefaultMaxTimeBetweenTransactions   = 200 * time.Second
	DefaultSendTransactionTimeout       = 10 * time.Second
	DefaultMetadataBackupPeriod         = 1 * time.Minute
	DefaultMaxInputGroupsForCreatingTx  = 4
	DefaultMaxOutputGroupsForCreatingTx = uint(4)
)

type BotConfig struct { //nolint:revive // BotConfig is a configuration struct for the bot.
	KeyPath                      string        `mapstructure:"key-path"`
	MetadataPath                 string        `mapstructure:"metadata-path"`
	MaxConcurrentAccounts        uint          `mapstructure:"max-concurrent-accounts"`
	MaxWalletsPerAccount         uint          `mapstructure:"max-wallets-per-account"`
	MinimumTxBalance             uint          `mapstructure:"minimum-tx-balance"`
	MinStartupFundDistribution   time.Duration `mapstructure:"min-startup-fund-distribution"`
	MaxStartupFundDistribution   time.Duration `mapstructure:"max-startup-fund-distribution"`
	MinTimeBetweenTransactions   time.Duration `mapstructure:"min-time-between-transactions"`
	MaxTimeBetweenTransactions   time.Duration `mapstructure:"max-time-between-transactions"`
	SendTransactionTimeout       time.Duration `mapstructure:"send-transaction-timeout"`
	MetadataBackupPeriod         time.Duration `mapstructure:"metadata-backup-period"`
	MaxInputGroupsForCreatingTx  int           `mapstructure:"max-input-groups-for-creating-tx"`
	MaxOutputGroupsForCreatingTx uint          `mapstructure:"max-output-groups-for-creating-tx"`
}

type Config struct {
	Logger *logrus.Logger
	Wallet walletcommon.ClientConfig `mapstructure:"wallet"`
	Bot    BotConfig                 `mapstructure:"bot"`
}

func NewConfig() *Config {
	return &Config{
		Logger: logrus.New(),
		Wallet: walletcommon.ClientConfig{
			ServerAddress:  walletcommon.DefaultServerAddress,
			ServerPort:     walletcommon.DefaultServerPort,
			RequestTimeout: walletcommon.DefaultRequestTimeout,
		},
		Bot: BotConfig{
			KeyPath:                      DefaultKeyPath,
			MetadataPath:                 DefaultMetadataPath,
			MaxConcurrentAccounts:        DefaultMaxConcurrentAccounts,
			MaxWalletsPerAccount:         DefaultMaxWalletsPerAccount,
			MinimumTxBalance:             DefaultMinimumTxBalance,
			MinStartupFundDistribution:   DefaultMinStartupFundDistribution,
			MaxStartupFundDistribution:   DefaultMaxStartupFundDistribution,
			MinTimeBetweenTransactions:   DefaultMinTimeBetweenTransactions,
			MaxTimeBetweenTransactions:   DefaultMaxTimeBetweenTransactions,
			SendTransactionTimeout:       DefaultSendTransactionTimeout,
			MetadataBackupPeriod:         DefaultMetadataBackupPeriod,
			MaxInputGroupsForCreatingTx:  DefaultMaxInputGroupsForCreatingTx,
			MaxOutputGroupsForCreatingTx: DefaultMaxOutputGroupsForCreatingTx,
		},
	}
}

func LoadConfig(cfgFile string) (*Config, error) {
	if cfgFile == "" {
		return nil, fmt.Errorf("config file not specified")
	}

	viper.SetConfigFile(cfgFile)
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	cfg := NewConfig()
	if err := viper.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("unable to decode into struct: %w", err)
	}

	cfg.Logger = logrus.New()

	return cfg, nil
}

func InitConfig(cmd *cobra.Command) *Config {
	cfg, err := LoadConfig(GetConfigFilePath(cmd))
	if err != nil {
		cfg = NewConfig()

		cfg.Logger.Infof("unable to load config file: %v", err)
		cfg.Logger.Infof("relying on default bot configuration")
	}

	ApplyFlagsToConfig(cmd, cfg)

	return cfg
}

func AddConfigFlags(cmd *cobra.Command) {
	cmd.Flags().String(KeyConfigFile, DefaultConfigFile, "config file (default is $PWD/config.yaml)")
	cmd.Flags().String(KeyWalletServerAddress, walletcommon.DefaultServerAddress, "Server address for wallet API requests")
	cmd.Flags().Uint(KeyWalletServerPort, walletcommon.DefaultServerPort, "Server port for wallet API requests")
	cmd.Flags().Duration(KeyWalletRequestTimeout, walletcommon.DefaultRequestTimeout, "Timeout for wallet API requests")

	cmd.Flags().String(KeyBotKeyPath, DefaultKeyPath, "Path to the HD wallet key")
	cmd.Flags().String(KeyBotMetadataPath, DefaultMetadataPath, "Path to the HD wallet metadata")
	cmd.Flags().Uint(KeyBotMaxConcurrentAccounts, DefaultMaxConcurrentAccounts, "Maximum number of concurrent bot accounts")
	cmd.Flags().Uint(KeyBotMaxWalletsPerAccount, DefaultMaxWalletsPerAccount, "Maximum number of wallets per bot account")
	cmd.Flags().Uint(KeyBotMinimumTxBalance, DefaultMinimumTxBalance, "Minimum balance used to decide whether a bot transaction should split outputs")
	cmd.Flags().Duration(KeyBotMinStartupFundDistribution, DefaultMinStartupFundDistribution, "Minimum startup delay before distributing funds")
	cmd.Flags().Duration(KeyBotMaxStartupFundDistribution, DefaultMaxStartupFundDistribution, "Maximum startup delay before distributing funds")
	cmd.Flags().Duration(KeyBotMinTimeBetweenTransactions, DefaultMinTimeBetweenTransactions, "Minimum delay between bot transactions")
	cmd.Flags().Duration(KeyBotMaxTimeBetweenTransactions, DefaultMaxTimeBetweenTransactions, "Maximum delay between bot transactions")
	cmd.Flags().Duration(KeyBotSendTransactionTimeout, DefaultSendTransactionTimeout, "Timeout for sending bot transactions")
	cmd.Flags().Duration(KeyBotMetadataBackupPeriod, DefaultMetadataBackupPeriod, "Period for saving bot HD wallet metadata")
	cmd.Flags().Int(KeyBotMaxInputGroupsForCreatingTx, DefaultMaxInputGroupsForCreatingTx, "Maximum number of input groups used when creating a bot transaction")
	cmd.Flags().Uint(KeyBotMaxOutputGroupsForCreatingTx, DefaultMaxOutputGroupsForCreatingTx, "Maximum number of output groups used when creating a bot transaction")

	_ = viper.BindPFlag(KeyConfigFile, cmd.Flags().Lookup(KeyConfigFile))
	_ = viper.BindPFlag(KeyWalletServerAddress, cmd.Flags().Lookup(KeyWalletServerAddress))
	_ = viper.BindPFlag(KeyWalletServerPort, cmd.Flags().Lookup(KeyWalletServerPort))
	_ = viper.BindPFlag(KeyWalletRequestTimeout, cmd.Flags().Lookup(KeyWalletRequestTimeout))

	_ = viper.BindPFlag(KeyBotKeyPath, cmd.Flags().Lookup(KeyBotKeyPath))
	_ = viper.BindPFlag(KeyBotMetadataPath, cmd.Flags().Lookup(KeyBotMetadataPath))
	_ = viper.BindPFlag(KeyBotMaxConcurrentAccounts, cmd.Flags().Lookup(KeyBotMaxConcurrentAccounts))
	_ = viper.BindPFlag(KeyBotMaxWalletsPerAccount, cmd.Flags().Lookup(KeyBotMaxWalletsPerAccount))
	_ = viper.BindPFlag(KeyBotMinimumTxBalance, cmd.Flags().Lookup(KeyBotMinimumTxBalance))
	_ = viper.BindPFlag(KeyBotMinStartupFundDistribution, cmd.Flags().Lookup(KeyBotMinStartupFundDistribution))
	_ = viper.BindPFlag(KeyBotMaxStartupFundDistribution, cmd.Flags().Lookup(KeyBotMaxStartupFundDistribution))
	_ = viper.BindPFlag(KeyBotMinTimeBetweenTransactions, cmd.Flags().Lookup(KeyBotMinTimeBetweenTransactions))
	_ = viper.BindPFlag(KeyBotMaxTimeBetweenTransactions, cmd.Flags().Lookup(KeyBotMaxTimeBetweenTransactions))
	_ = viper.BindPFlag(KeyBotSendTransactionTimeout, cmd.Flags().Lookup(KeyBotSendTransactionTimeout))
	_ = viper.BindPFlag(KeyBotMetadataBackupPeriod, cmd.Flags().Lookup(KeyBotMetadataBackupPeriod))
	_ = viper.BindPFlag(KeyBotMaxInputGroupsForCreatingTx, cmd.Flags().Lookup(KeyBotMaxInputGroupsForCreatingTx))
	_ = viper.BindPFlag(KeyBotMaxOutputGroupsForCreatingTx, cmd.Flags().Lookup(KeyBotMaxOutputGroupsForCreatingTx))
}

func GetConfigFilePath(cmd *cobra.Command) string {
	if cmd.Flags().Changed(KeyConfigFile) {
		return viper.GetString(KeyConfigFile)
	}
	return ""
}

func ApplyFlagsToConfig(cmd *cobra.Command, cfg *Config) {
	if cmd.Flags().Changed(KeyWalletServerAddress) {
		cfg.Wallet.ServerAddress = viper.GetString(KeyWalletServerAddress)
	}
	if cmd.Flags().Changed(KeyWalletServerPort) {
		cfg.Wallet.ServerPort = viper.GetUint(KeyWalletServerPort)
	}
	if cmd.Flags().Changed(KeyWalletRequestTimeout) {
		cfg.Wallet.RequestTimeout = viper.GetDuration(KeyWalletRequestTimeout)
	}
	if cmd.Flags().Changed(KeyBotKeyPath) {
		cfg.Bot.KeyPath = viper.GetString(KeyBotKeyPath)
	}
	if cmd.Flags().Changed(KeyBotMetadataPath) {
		cfg.Bot.MetadataPath = viper.GetString(KeyBotMetadataPath)
	}
	if cmd.Flags().Changed(KeyBotMaxConcurrentAccounts) {
		cfg.Bot.MaxConcurrentAccounts = viper.GetUint(KeyBotMaxConcurrentAccounts)
	}
	if cmd.Flags().Changed(KeyBotMaxWalletsPerAccount) {
		cfg.Bot.MaxWalletsPerAccount = viper.GetUint(KeyBotMaxWalletsPerAccount)
	}
	if cmd.Flags().Changed(KeyBotMinimumTxBalance) {
		cfg.Bot.MinimumTxBalance = viper.GetUint(KeyBotMinimumTxBalance)
	}
	if cmd.Flags().Changed(KeyBotMinStartupFundDistribution) {
		cfg.Bot.MinStartupFundDistribution = viper.GetDuration(KeyBotMinStartupFundDistribution)
	}
	if cmd.Flags().Changed(KeyBotMaxStartupFundDistribution) {
		cfg.Bot.MaxStartupFundDistribution = viper.GetDuration(KeyBotMaxStartupFundDistribution)
	}
	if cmd.Flags().Changed(KeyBotMinTimeBetweenTransactions) {
		cfg.Bot.MinTimeBetweenTransactions = viper.GetDuration(KeyBotMinTimeBetweenTransactions)
	}
	if cmd.Flags().Changed(KeyBotMaxTimeBetweenTransactions) {
		cfg.Bot.MaxTimeBetweenTransactions = viper.GetDuration(KeyBotMaxTimeBetweenTransactions)
	}
	if cmd.Flags().Changed(KeyBotSendTransactionTimeout) {
		cfg.Bot.SendTransactionTimeout = viper.GetDuration(KeyBotSendTransactionTimeout)
	}
	if cmd.Flags().Changed(KeyBotMetadataBackupPeriod) {
		cfg.Bot.MetadataBackupPeriod = viper.GetDuration(KeyBotMetadataBackupPeriod)
	}
	if cmd.Flags().Changed(KeyBotMaxInputGroupsForCreatingTx) {
		cfg.Bot.MaxInputGroupsForCreatingTx = viper.GetInt(KeyBotMaxInputGroupsForCreatingTx)
	}
	if cmd.Flags().Changed(KeyBotMaxOutputGroupsForCreatingTx) {
		cfg.Bot.MaxOutputGroupsForCreatingTx = viper.GetUint(KeyBotMaxOutputGroupsForCreatingTx)
	}
}
