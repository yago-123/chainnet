package config

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// default config keys
const (
	KeyConfigFile               = "config"
	KeyNodeSeeds                = "node-seeds"
	KeyStorageFile              = "storage-file"
	KeyPubKey                   = "pub-key"
	KeyMiningInterval           = "mining-interval"
	KeyTargetIntervalAdjustment = "block-interval-adjustment"
	KeyP2PEnabled               = "enabled"
	KeyP2PPeerPrivKey           = "priv-key-path"
	KeyP2PPeerPort              = "peer-port"
	KeyP2PMinNumConn            = "min-conn"
	KeyP2PMaxNumConn            = "max-conn"
	KeyP2PConnTimeout           = "conn-timeout"
	KeyP2PWriteTimeout          = "write-timeout"
	KeyP2PReadTimeout           = "read-timeout"
	KeyP2PBufferSize            = "buffer-size"
)

// default config values
const (
	DefaultConfigFile = ""

	DefaultChainnetStorage          = "chainnet-storage"
	DefaultMiningInterval           = 10 * time.Minute
	DefaultTargetIntervalAdjustment = uint(6)
	DefaultP2PEnabled               = true
	DefaultP2PPeerPort              = 9100
	DefaultP2PMinNumConn            = 1
	DefaultP2PMaxNumConn            = 100
	DefaultP2PConnTimeout           = 20 * time.Second
	DefaultP2PWriteTimeout          = 10 * time.Second
	DefaultP2PReadTimeout           = 10 * time.Second
	DefaultP2PBufferSize            = 8192
)

const (
	SeedNodeNumberArguments = 3
)

// SeedNode represents a node in the configuration with address, peerID, and port
type SeedNode struct {
	Address string `mapstructure:"address"`
	PeerID  string `mapstructure:"peer-id"`
	Port    int    `mapstructure:"port"`
}

// IdentityConfig holds the identity-specific configuration
type IdentityConfig struct {
	PrivKeyPath string `mapstructure:"priv-key-path"`
}

// P2PConfig holds P2P-specific configuration
type P2PConfig struct {
	Enabled      bool           `mapstructure:"enabled"`
	Identity     IdentityConfig `mapstructure:"identity"`
	PeerPort     uint           `mapstructure:"peer-port"`
	RouterPort   uint           `mapstructure:"http-api-port"`
	MinNumConn   uint           `mapstructure:"min-conn"`
	MaxNumConn   uint           `mapstructure:"max-conn"`
	ConnTimeout  time.Duration  `mapstructure:"conn-timeout"`
	WriteTimeout time.Duration  `mapstructure:"write-timeout"`
	ReadTimeout  time.Duration  `mapstructure:"read-timeout"`
	BufferSize   uint           `mapstructure:"buffer-size"`
}

// Config holds the configuration for the application
type Config struct {
	Logger                   *logrus.Logger
	SeedNodes                []SeedNode    `mapstructure:"seed-nodes"`
	StorageFile              string        `mapstructure:"storage-file"`
	PubKey                   string        `mapstructure:"pub-key"`
	MiningInterval           time.Duration `mapstructure:"mining-interval"`
	AdjustmentTargetInterval uint          `mapstructure:"block-interval-adjustment"`
	P2P                      P2PConfig     `mapstructure:"p2p"`
}

// NewConfig creates a new Config with default values
func NewConfig() *Config {
	return &Config{
		Logger:                   logrus.New(),
		SeedNodes:                []SeedNode{},
		StorageFile:              DefaultChainnetStorage,
		PubKey:                   "",
		MiningInterval:           DefaultMiningInterval,
		AdjustmentTargetInterval: DefaultTargetIntervalAdjustment,
		P2P: P2PConfig{
			Enabled: DefaultP2PEnabled,
			Identity: IdentityConfig{
				PrivKeyPath: "",
			},
			MinNumConn:   DefaultP2PMinNumConn,
			MaxNumConn:   DefaultP2PMaxNumConn,
			ConnTimeout:  DefaultP2PConnTimeout,
			WriteTimeout: DefaultP2PWriteTimeout,
			ReadTimeout:  DefaultP2PReadTimeout,
			BufferSize:   DefaultP2PBufferSize,
		},
	}
}

// LoadConfig loads configuration from the specified file
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

// InitConfig initializes configuration, loading from file and applying flags
func InitConfig(cmd *cobra.Command) *Config {
	cfg, err := LoadConfig(GetConfigFilePath(cmd))
	if err != nil {
		cfg = NewConfig()

		cfg.Logger.Infof("unable to load config file: %v", err)
		cfg.Logger.Infof("relying on default configuration")
	}

	ApplyFlagsToConfig(cmd, cfg)

	return cfg
}

// AddConfigFlags adds flags for configuration options to the command
func AddConfigFlags(cmd *cobra.Command) {
	// define flags
	cmd.Flags().String(KeyConfigFile, DefaultConfigFile, "config file (default is $PWD/config.yaml)")
	cmd.Flags().StringArray(KeyNodeSeeds, []string{}, "Node seeds used to synchronize during startup")
	cmd.Flags().String(KeyStorageFile, DefaultChainnetStorage, "Storage file name")
	cmd.Flags().String(KeyPubKey, "", "Public key used for receiving mining rewards")
	cmd.Flags().Duration(KeyMiningInterval, DefaultMiningInterval, "Mining interval in seconds")
	cmd.Flags().Uint(KeyTargetIntervalAdjustment, DefaultTargetIntervalAdjustment, "Number of blocks for adjusting difficulty")
	cmd.Flags().Bool(KeyP2PEnabled, DefaultP2PEnabled, "Enable P2P")
	cmd.Flags().String(KeyP2PPeerPrivKey, "", "ECDSA peer private key path in PEM format")
	cmd.Flags().Uint(KeyP2PPeerPort, DefaultP2PPeerPort, "Peer port")
	cmd.Flags().Uint(KeyP2PMinNumConn, DefaultP2PMinNumConn, "Minimum number of P2P connections")
	cmd.Flags().Uint(KeyP2PMaxNumConn, DefaultP2PMaxNumConn, "Maximum number of P2P connections")
	cmd.Flags().Duration(KeyP2PConnTimeout, DefaultP2PConnTimeout, "P2P connection timeout")
	cmd.Flags().Duration(KeyP2PWriteTimeout, DefaultP2PWriteTimeout, "P2P write timeout")
	cmd.Flags().Duration(KeyP2PReadTimeout, DefaultP2PReadTimeout, "P2P read timeout")
	cmd.Flags().Uint(KeyP2PBufferSize, DefaultP2PBufferSize, "P2P buffer size for reading from stream")

	// bind flags to viper
	_ = viper.BindPFlag(KeyConfigFile, cmd.Flags().Lookup(KeyConfigFile))
	_ = viper.BindPFlag(KeyNodeSeeds, cmd.Flags().Lookup(KeyNodeSeeds))
	_ = viper.BindPFlag(KeyStorageFile, cmd.Flags().Lookup(KeyStorageFile))
	_ = viper.BindPFlag(KeyPubKey, cmd.Flags().Lookup(KeyPubKey))
	_ = viper.BindPFlag(KeyMiningInterval, cmd.Flags().Lookup(KeyMiningInterval))
	_ = viper.BindPFlag(KeyTargetIntervalAdjustment, cmd.Flags().Lookup(KeyTargetIntervalAdjustment))
	_ = viper.BindPFlag(KeyP2PEnabled, cmd.Flags().Lookup(KeyP2PEnabled))
	_ = viper.BindPFlag(KeyP2PPeerPrivKey, cmd.Flags().Lookup(KeyP2PPeerPrivKey))
	_ = viper.BindPFlag(KeyP2PPeerPort, cmd.Flags().Lookup(KeyP2PPeerPort))
	_ = viper.BindPFlag(KeyP2PMinNumConn, cmd.Flags().Lookup(KeyP2PMinNumConn))
	_ = viper.BindPFlag(KeyP2PMaxNumConn, cmd.Flags().Lookup(KeyP2PMaxNumConn))
	_ = viper.BindPFlag(KeyP2PConnTimeout, cmd.Flags().Lookup(KeyP2PConnTimeout))
	_ = viper.BindPFlag(KeyP2PWriteTimeout, cmd.Flags().Lookup(KeyP2PWriteTimeout))
	_ = viper.BindPFlag(KeyP2PReadTimeout, cmd.Flags().Lookup(KeyP2PReadTimeout))
	_ = viper.BindPFlag(KeyP2PBufferSize, cmd.Flags().Lookup(KeyP2PBufferSize))
}

// GetConfigFilePath retrieves the configuration file path from command flags
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
		nodeSeeds := viper.GetStringSlice(KeyNodeSeeds)
		seeds, err := parseSeedNodes(nodeSeeds)
		if err != nil {
			cfg.Logger.Errorf("error parsing seed nodes: %v", err)
		} else {
			cfg.SeedNodes = seeds
		}
	}
	if cmd.Flags().Changed(KeyStorageFile) {
		cfg.StorageFile = viper.GetString(KeyStorageFile)
	}
	if cmd.Flags().Changed(KeyPubKey) {
		cfg.PubKey = viper.GetString(KeyPubKey)
	}
	if cmd.Flags().Changed(KeyMiningInterval) {
		cfg.MiningInterval = viper.GetDuration(KeyMiningInterval)
	}
	if cmd.Flags().Changed(KeyTargetIntervalAdjustment) {
		cfg.AdjustmentTargetInterval = viper.GetUint(KeyTargetIntervalAdjustment)
	}
	if cmd.Flags().Changed(KeyP2PEnabled) {
		cfg.P2P.Enabled = viper.GetBool(KeyP2PEnabled)
	}
	if cmd.Flags().Changed(KeyP2PPeerPrivKey) {
		cfg.P2P.Identity.PrivKeyPath = viper.GetString(KeyP2PPeerPrivKey)
	}
	if cmd.Flags().Changed(KeyP2PPeerPort) {
		cfg.P2P.PeerPort = viper.GetUint(KeyP2PPeerPort)
	}
	if cmd.Flags().Changed(KeyP2PMinNumConn) {
		cfg.P2P.MinNumConn = viper.GetUint(KeyP2PMinNumConn)
	}
	if cmd.Flags().Changed(KeyP2PMaxNumConn) {
		cfg.P2P.MaxNumConn = viper.GetUint(KeyP2PMaxNumConn)
	}
	if cmd.Flags().Changed(KeyP2PConnTimeout) {
		cfg.P2P.ConnTimeout = viper.GetDuration(KeyP2PConnTimeout)
	}
	if cmd.Flags().Changed(KeyP2PWriteTimeout) {
		cfg.P2P.WriteTimeout = viper.GetDuration(KeyP2PWriteTimeout)
	}
	if cmd.Flags().Changed(KeyP2PReadTimeout) {
		cfg.P2P.ReadTimeout = viper.GetDuration(KeyP2PReadTimeout)
	}
	if cmd.Flags().Changed(KeyP2PBufferSize) {
		cfg.P2P.BufferSize = viper.GetUint(KeyP2PBufferSize)
	}
}

// parseSeedNodes parses seed nodes from a slice of strings and returns a slice of SeedNode structs
func parseSeedNodes(seedNodes []string) ([]SeedNode, error) {
	var seeds []SeedNode
	for _, nodeSeed := range seedNodes {
		parts := strings.SplitN(nodeSeed, ":", SeedNodeNumberArguments)
		if len(parts) == SeedNodeNumberArguments {
			// make sure that seed nodes have all the fields required
			port, err := strconv.Atoi(parts[2])
			if err != nil {
				return nil, err
			}
			seeds = append(seeds, SeedNode{Address: parts[0], PeerID: parts[1], Port: port})
		} else if len(parts) != SeedNodeNumberArguments {
			// otherwise return an error
			return nil, fmt.Errorf("invalid seed node format: %s", nodeSeed)
		}
	}
	return seeds, nil
}
