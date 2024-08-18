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
	KeyConfigFile      = "config"
	KeyNodeSeeds       = "node-seeds"
	KeyStorageFile     = "storage-file"
	KeyMiningInterval  = "mining-interval"
	KeyP2PEnabled      = "p2p-enabled"
	KeyP2PMinNumConn   = "p2p-min-conn"
	KeyP2PMaxNumConn   = "p2p-max-conn"
	KeyP2PConnTimeout  = "p2p-conn-timeout"
	KeyP2PWriteTimeout = "p2p-write-timeout" //nolint:gosec // no hardcoded credential
	KeyP2PReadTimeout  = "p2p-read-timeout"
	KeyP2PBufferSize   = "p2p-buffer-size"
)

// default config values
const (
	DefaultConfigFile = ""

	DefaultNodeSeed  = "seed.chainnet.yago.ninja"
	DefaultNodeSeed2 = "seed-2.chainnet.yago.ninja"
	DefaultNodeSeed3 = "seed-3.chainnet.yago.ninja"

	DefaultChainnetStorage = "chainnet-storage"
	DefaultMiningInterval  = 1 * time.Minute
	DefaultP2PEnabled      = true
	DefaultP2PMinNumConn   = 1
	DefaultP2PMaxNumConn   = 100
	DefaultP2PConnTimeout  = 20 * time.Second
	DefaultP2PWriteTimeout = 10 * time.Second
	DefaultP2PReadTimeout  = 10 * time.Second
	DefaultP2PBufferSize   = 8192
)

type Config struct {
	Logger          *logrus.Logger
	NodeSeeds       []string      `mapstructure:"node-seeds"`
	StorageFile     string        `mapstructure:"storage-file"`
	MiningInterval  time.Duration `mapstructure:"mining-interval"`
	P2PEnabled      bool          `mapstructure:"p2p-enabled"`
	P2PMinNumConn   uint          `mapstructure:"p2p-min-conn"`
	P2PMaxNumConn   uint          `mapstructure:"p2p-max-conn"`
	P2PConnTimeout  time.Duration `mapstructure:"p2p-conn-timeout"`
	P2PWriteTimeout time.Duration `mapstructure:"p2p-write-timeout"`
	P2PReadTimeout  time.Duration `mapstructure:"p2p-read-timeout"`
	P2PBufferSize   uint          `mapstructure:"p2p-buffer-size"`
}

func NewConfig() *Config {
	return &Config{
		Logger:          logrus.New(),
		NodeSeeds:       []string{DefaultNodeSeed, DefaultNodeSeed2, DefaultNodeSeed3},
		StorageFile:     DefaultChainnetStorage,
		MiningInterval:  DefaultMiningInterval,
		P2PEnabled:      DefaultP2PEnabled,
		P2PMinNumConn:   DefaultP2PMinNumConn,
		P2PMaxNumConn:   DefaultP2PMaxNumConn,
		P2PConnTimeout:  DefaultP2PConnTimeout,
		P2PWriteTimeout: DefaultP2PWriteTimeout,
		P2PReadTimeout:  DefaultP2PReadTimeout,
		P2PBufferSize:   DefaultP2PBufferSize,
	}
}

func (c *Config) SetP2PStatus(enable bool) {
	c.P2PEnabled = enable
}

func LoadConfig(cfgFile string) (*Config, error) {
	var cfg Config

	if cfgFile == "" {
		return nil, fmt.Errorf("config file not specified")
	}

	viper.SetConfigFile(cfgFile)
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unable to decode into struct: %w", err)
	}

	cfg.Logger = logrus.New()

	return &cfg, nil
}

func InitConfig(cmd *cobra.Command) *Config {
	cfg, err := LoadConfig(GetConfigFilePath(cmd))
	if err != nil {
		cfg = NewConfig()

		cfg.Logger.Infof("unable to load config file: %v", err)
		cfg.Logger.Infof("relying in default configuration")
	}

	ApplyFlagsToConfig(cmd, cfg)

	return cfg
}

func AddConfigFlags(cmd *cobra.Command) {
	cmd.Flags().String(KeyConfigFile, DefaultConfigFile, "config file (default is $PWD/config.yaml)")
	cmd.Flags().StringArray(KeyNodeSeeds, []string{DefaultNodeSeed, DefaultNodeSeed2, DefaultNodeSeed3}, "Node seeds used to synchronize during startup")
	cmd.Flags().String(KeyStorageFile, DefaultChainnetStorage, "Storage file name")
	cmd.Flags().Duration(KeyMiningInterval, DefaultMiningInterval, "Mining interval in seconds")
	cmd.Flags().Bool(KeyP2PEnabled, DefaultP2PEnabled, "Enable P2P")
	cmd.Flags().Uint(KeyP2PMinNumConn, DefaultP2PMinNumConn, "Minimum number of P2P connections")
	cmd.Flags().Uint(KeyP2PMaxNumConn, DefaultP2PMaxNumConn, "Maximum number of P2P connections")
	cmd.Flags().Duration(KeyP2PConnTimeout, DefaultP2PConnTimeout, "P2P connection timeout")
	cmd.Flags().Duration(KeyP2PWriteTimeout, DefaultP2PWriteTimeout, "P2P write timeout")
	cmd.Flags().Duration(KeyP2PReadTimeout, DefaultP2PReadTimeout, "P2P read timeout")
	cmd.Flags().Uint(KeyP2PBufferSize, DefaultP2PBufferSize, "P2P buffer size for reading from stream")

	_ = viper.BindPFlag(KeyConfigFile, cmd.Flags().Lookup(KeyConfigFile))
	_ = viper.BindPFlag(KeyStorageFile, cmd.Flags().Lookup(KeyStorageFile))
	_ = viper.BindPFlag(KeyMiningInterval, cmd.Flags().Lookup(KeyMiningInterval))
	_ = viper.BindPFlag(KeyP2PEnabled, cmd.Flags().Lookup(KeyP2PEnabled))
	_ = viper.BindPFlag(KeyP2PMinNumConn, cmd.Flags().Lookup(KeyP2PMinNumConn))
	_ = viper.BindPFlag(KeyP2PMaxNumConn, cmd.Flags().Lookup(KeyP2PMaxNumConn))
	_ = viper.BindPFlag(KeyP2PConnTimeout, cmd.Flags().Lookup(KeyP2PConnTimeout))
	_ = viper.BindPFlag(KeyP2PWriteTimeout, cmd.Flags().Lookup(KeyP2PWriteTimeout))
	_ = viper.BindPFlag(KeyP2PReadTimeout, cmd.Flags().Lookup(KeyP2PReadTimeout))
	_ = viper.BindPFlag(KeyP2PBufferSize, cmd.Flags().Lookup(KeyP2PBufferSize))
}

func GetConfigFilePath(cmd *cobra.Command) string {
	if cmd.Flags().Changed(KeyConfigFile) {
		return viper.GetString(KeyConfigFile)
	}

	return ""
}

// ApplyFlagsToConfig updates the config struct with flag values if they have been set
func ApplyFlagsToConfig(cmd *cobra.Command, cfg *Config) {
	// todo(): use flag-to-config mapping function
	if cmd.Flags().Changed(KeyNodeSeeds) {
		cfg.NodeSeeds = viper.GetStringSlice(KeyNodeSeeds)
	}
	if cmd.Flags().Changed(KeyStorageFile) {
		cfg.StorageFile = viper.GetString(KeyStorageFile)
	}
	if cmd.Flags().Changed(KeyMiningInterval) {
		cfg.MiningInterval = viper.GetDuration(KeyMiningInterval)
	}
	if cmd.Flags().Changed(KeyP2PEnabled) {
		cfg.P2PEnabled = viper.GetBool(KeyP2PEnabled)
	}
	if cmd.Flags().Changed(KeyP2PMinNumConn) {
		cfg.P2PMinNumConn = viper.GetUint(KeyP2PMinNumConn)
	}
	if cmd.Flags().Changed(KeyP2PMaxNumConn) {
		cfg.P2PMaxNumConn = viper.GetUint(KeyP2PMaxNumConn)
	}
	if cmd.Flags().Changed(KeyP2PConnTimeout) {
		cfg.P2PConnTimeout = viper.GetDuration(KeyP2PConnTimeout)
	}
	if cmd.Flags().Changed(KeyP2PWriteTimeout) {
		cfg.P2PWriteTimeout = viper.GetDuration(KeyP2PWriteTimeout)
	}
	if cmd.Flags().Changed(KeyP2PReadTimeout) {
		cfg.P2PReadTimeout = viper.GetDuration(KeyP2PReadTimeout)
	}
	if cmd.Flags().Changed(KeyP2PBufferSize) {
		cfg.P2PBufferSize = viper.GetUint(KeyP2PBufferSize)
	}
}
