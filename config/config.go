package config

import (
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// default config keys
const (
	KeyStorageFile    = "storage-file"
	KeyMiningInterval = "mining-interval"
	KeyP2PEnabled     = "p2p-enabled"
	KeyP2PMinNumConn  = "min-p2p-conn"
	KeyP2PMaxNumConn  = "max-p2p-conn"
)

// default config values
const (
	DefaultChainnetStorage = "chainnet-storage"
	DefaultMiningInterval  = 1 * time.Minute
	DefaultP2PEnabled      = true
	DefaultP2PMinNumConn   = 1
	DefaultP2PMaxNumConn   = 100
)

type Config struct {
	Logger         *logrus.Logger
	StorageFile    string
	MiningInterval time.Duration
	P2PEnabled     bool
	P2PMinNumConn  uint
	P2PMaxNumConn  uint
}

func NewConfig() *Config {
	return &Config{
		Logger:         logrus.New(),
		StorageFile:    DefaultChainnetStorage,
		MiningInterval: DefaultMiningInterval,
		P2PEnabled:     DefaultP2PEnabled,
		P2PMinNumConn:  DefaultP2PMinNumConn,
		P2PMaxNumConn:  DefaultP2PMaxNumConn,
	}
}

func (c *Config) SetP2PStatus(enable bool) {
	c.P2PEnabled = enable
}

func LoadConfig(cfgFile string) (*Config, error) {
	cfg := NewConfig()

	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
		viper.AddConfigPath(".")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	if err := viper.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("unable to decode into struct: %w", err)
	}

	return cfg, nil
}

func AddConfigFlags(cmd *cobra.Command) {
	cmd.Flags().String(KeyStorageFile, DefaultChainnetStorage, "Storage file name")
	cmd.Flags().Duration(KeyMiningInterval, DefaultMiningInterval, "Mining interval in seconds")
	cmd.Flags().Uint(KeyP2PMinNumConn, DefaultP2PMinNumConn, "Minimum number of P2P connections")
	cmd.Flags().Uint(KeyP2PMaxNumConn, DefaultP2PMaxNumConn, "Maximum number of P2P connections")

	_ = viper.BindPFlag(KeyStorageFile, cmd.Flags().Lookup(KeyStorageFile))
	_ = viper.BindPFlag(KeyMiningInterval, cmd.Flags().Lookup(KeyMiningInterval))
	_ = viper.BindPFlag(KeyP2PMinNumConn, cmd.Flags().Lookup(KeyP2PMinNumConn))
	_ = viper.BindPFlag(KeyP2PMaxNumConn, cmd.Flags().Lookup(KeyP2PMaxNumConn))
}

// ApplyFlagsToConfig updates the config struct with flag values if they have been set
func ApplyFlagsToConfig(cmd *cobra.Command, cfg *Config) {
	if cmd.Flags().Changed(KeyStorageFile) {
		cfg.StorageFile = viper.GetString(KeyStorageFile)
	}

	if cmd.Flags().Changed(KeyMiningInterval) {
		cfg.MiningInterval = viper.GetDuration(KeyMiningInterval)
	}
	if cmd.Flags().Changed(KeyP2PMinNumConn) {
		cfg.P2PMinNumConn = viper.GetUint(KeyP2PMinNumConn)
	}
	if cmd.Flags().Changed(KeyP2PMaxNumConn) {
		cfg.P2PMaxNumConn = viper.GetUint(KeyP2PMaxNumConn)
	}
}
