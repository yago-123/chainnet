package bot

import (
	"fmt"
	"strings"
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

	v := viper.New()
	v.SetConfigFile(cfgFile)
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	cfg := NewConfig()
	if err := v.Unmarshal(cfg); err != nil {
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
	if err := cfg.Validate(); err != nil {
		cfg.Logger.Fatalf("invalid bot configuration: %v", err)
	}

	return cfg
}

func (cfg *Config) Validate() error {
	if cfg == nil {
		return fmt.Errorf("config is nil")
	}

	var violations []string

	if strings.TrimSpace(cfg.Wallet.ServerAddress) == "" {
		violations = append(violations, "wallet server address must not be empty")
	}
	if cfg.Wallet.ServerPort == 0 {
		violations = append(violations, "wallet server port must be greater than 0")
	}
	if cfg.Wallet.RequestTimeout <= 0 {
		violations = append(violations, "wallet request timeout must be greater than 0")
	}

	if strings.TrimSpace(cfg.Bot.KeyPath) == "" {
		violations = append(violations, "bot key path must not be empty")
	}
	if strings.TrimSpace(cfg.Bot.MetadataPath) == "" {
		violations = append(violations, "bot metadata path must not be empty")
	}
	if cfg.Bot.MaxConcurrentAccounts == 0 {
		violations = append(violations, "bot max concurrent accounts must be greater than 0")
	}
	if cfg.Bot.MaxWalletsPerAccount == 0 {
		violations = append(violations, "bot max wallets per account must be greater than 0")
	}
	if cfg.Bot.MinStartupFundDistribution < 0 {
		violations = append(violations, "bot min startup fund distribution must not be negative")
	}
	if cfg.Bot.MaxStartupFundDistribution < 0 {
		violations = append(violations, "bot max startup fund distribution must not be negative")
	}
	if cfg.Bot.MaxStartupFundDistribution < cfg.Bot.MinStartupFundDistribution {
		violations = append(violations, "bot max startup fund distribution must be greater than or equal to min startup fund distribution")
	}
	if cfg.Bot.MinTimeBetweenTransactions < 0 {
		violations = append(violations, "bot min time between transactions must not be negative")
	}
	if cfg.Bot.MaxTimeBetweenTransactions < 0 {
		violations = append(violations, "bot max time between transactions must not be negative")
	}
	if cfg.Bot.MaxTimeBetweenTransactions < cfg.Bot.MinTimeBetweenTransactions {
		violations = append(violations, "bot max time between transactions must be greater than or equal to min time between transactions")
	}
	if cfg.Bot.SendTransactionTimeout <= 0 {
		violations = append(violations, "bot send transaction timeout must be greater than 0")
	}
	if cfg.Bot.MetadataBackupPeriod <= 0 {
		violations = append(violations, "bot metadata backup period must be greater than 0")
	}
	if cfg.Bot.MaxInputGroupsForCreatingTx <= 0 {
		violations = append(violations, "bot max input groups for creating tx must be greater than 0")
	}
	if cfg.Bot.MaxOutputGroupsForCreatingTx == 0 {
		violations = append(violations, "bot max output groups for creating tx must be greater than 0")
	}
	if cfg.Bot.MaxOutputGroupsForCreatingTx > cfg.Bot.MaxWalletsPerAccount {
		violations = append(violations, "bot max output groups must be smaller or equal than bot max wallets per account")
	}

	if len(violations) > 0 {
		return fmt.Errorf("%s", strings.Join(violations, "; "))
	}

	return nil
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
}

func GetConfigFilePath(cmd *cobra.Command) string {
	if cmd.Flags().Changed(KeyConfigFile) {
		cfgFile, err := cmd.Flags().GetString(KeyConfigFile)
		if err != nil {
			return ""
		}

		return cfgFile
	}
	return ""
}

func ApplyFlagsToConfig(cmd *cobra.Command, cfg *Config) {
	if cmd.Flags().Changed(KeyWalletServerAddress) {
		cfg.Wallet.ServerAddress = mustGetStringFlag(cmd, KeyWalletServerAddress)
	}
	if cmd.Flags().Changed(KeyWalletServerPort) {
		cfg.Wallet.ServerPort = mustGetUintFlag(cmd, KeyWalletServerPort)
	}
	if cmd.Flags().Changed(KeyWalletRequestTimeout) {
		cfg.Wallet.RequestTimeout = mustGetDurationFlag(cmd, KeyWalletRequestTimeout)
	}
	if cmd.Flags().Changed(KeyBotKeyPath) {
		cfg.Bot.KeyPath = mustGetStringFlag(cmd, KeyBotKeyPath)
	}
	if cmd.Flags().Changed(KeyBotMetadataPath) {
		cfg.Bot.MetadataPath = mustGetStringFlag(cmd, KeyBotMetadataPath)
	}
	if cmd.Flags().Changed(KeyBotMaxConcurrentAccounts) {
		cfg.Bot.MaxConcurrentAccounts = mustGetUintFlag(cmd, KeyBotMaxConcurrentAccounts)
	}
	if cmd.Flags().Changed(KeyBotMaxWalletsPerAccount) {
		cfg.Bot.MaxWalletsPerAccount = mustGetUintFlag(cmd, KeyBotMaxWalletsPerAccount)
	}
	if cmd.Flags().Changed(KeyBotMinimumTxBalance) {
		cfg.Bot.MinimumTxBalance = mustGetUintFlag(cmd, KeyBotMinimumTxBalance)
	}
	if cmd.Flags().Changed(KeyBotMinStartupFundDistribution) {
		cfg.Bot.MinStartupFundDistribution = mustGetDurationFlag(cmd, KeyBotMinStartupFundDistribution)
	}
	if cmd.Flags().Changed(KeyBotMaxStartupFundDistribution) {
		cfg.Bot.MaxStartupFundDistribution = mustGetDurationFlag(cmd, KeyBotMaxStartupFundDistribution)
	}
	if cmd.Flags().Changed(KeyBotMinTimeBetweenTransactions) {
		cfg.Bot.MinTimeBetweenTransactions = mustGetDurationFlag(cmd, KeyBotMinTimeBetweenTransactions)
	}
	if cmd.Flags().Changed(KeyBotMaxTimeBetweenTransactions) {
		cfg.Bot.MaxTimeBetweenTransactions = mustGetDurationFlag(cmd, KeyBotMaxTimeBetweenTransactions)
	}
	if cmd.Flags().Changed(KeyBotSendTransactionTimeout) {
		cfg.Bot.SendTransactionTimeout = mustGetDurationFlag(cmd, KeyBotSendTransactionTimeout)
	}
	if cmd.Flags().Changed(KeyBotMetadataBackupPeriod) {
		cfg.Bot.MetadataBackupPeriod = mustGetDurationFlag(cmd, KeyBotMetadataBackupPeriod)
	}
	if cmd.Flags().Changed(KeyBotMaxInputGroupsForCreatingTx) {
		cfg.Bot.MaxInputGroupsForCreatingTx = mustGetIntFlag(cmd, KeyBotMaxInputGroupsForCreatingTx)
	}
	if cmd.Flags().Changed(KeyBotMaxOutputGroupsForCreatingTx) {
		cfg.Bot.MaxOutputGroupsForCreatingTx = mustGetUintFlag(cmd, KeyBotMaxOutputGroupsForCreatingTx)
	}
}

func mustGetStringFlag(cmd *cobra.Command, name string) string {
	value, err := cmd.Flags().GetString(name)
	if err != nil {
		panic(fmt.Sprintf("error reading %s flag: %v", name, err))
	}

	return value
}

func mustGetUintFlag(cmd *cobra.Command, name string) uint {
	value, err := cmd.Flags().GetUint(name)
	if err != nil {
		panic(fmt.Sprintf("error reading %s flag: %v", name, err))
	}

	return value
}

func mustGetIntFlag(cmd *cobra.Command, name string) int {
	value, err := cmd.Flags().GetInt(name)
	if err != nil {
		panic(fmt.Sprintf("error reading %s flag: %v", name, err))
	}

	return value
}

func mustGetDurationFlag(cmd *cobra.Command, name string) time.Duration {
	value, err := cmd.Flags().GetDuration(name)
	if err != nil {
		panic(fmt.Sprintf("error reading %s flag: %v", name, err))
	}

	return value
}
