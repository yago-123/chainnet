package config

import (
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	DefaultMiningInterval = 1 * time.Minute
	DefaultP2PEnabled     = true
	DefaultP2PMinNumConn  = 1
	DefaultP2PMaxNumConn  = 100
)

type Config struct {
	Logger         *logrus.Logger
	MiningInterval time.Duration
	P2PEnabled     bool
	P2PMinNumConn  uint
	P2PMaxNumConn  uint
}

func NewConfig() *Config {
	return &Config{
		Logger:         logrus.New(),
		MiningInterval: DefaultMiningInterval,
		P2PEnabled:     DefaultP2PEnabled,
		P2PMinNumConn:  DefaultP2PMinNumConn,
		P2PMaxNumConn:  DefaultP2PMaxNumConn,
	}
}

func (c *Config) SetP2PEnabled(enable bool) {
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
	cmd.Flags().Duration("mining-interval", DefaultMiningInterval, "Mining interval in seconds")
	cmd.Flags().Uint("min-num-p2p-conn", DefaultP2PMinNumConn, "Minimum number of P2P connections")
	cmd.Flags().Uint("max-num-p2p-conn", DefaultP2PMaxNumConn, "Maximum number of P2P connections")

	viper.BindPFlag("mining-interval", cmd.Flags().Lookup("mining-interval"))
	viper.BindPFlag("min-num-p2p-conn", cmd.Flags().Lookup("min-num-p2p-conn"))
	viper.BindPFlag("max-num-p2p-conn", cmd.Flags().Lookup("max-num-p2p-conn"))
}
