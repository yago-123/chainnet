package wallet

import (
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
)

const (
	DefaultServerAddress  = "seed-1.chainnet.yago.ninja"
	DefaultServerPort     = uint(8080)
	DefaultRequestTimeout = 20 * time.Second
)

type ClientConfig struct {
	ServerAddress  string         `mapstructure:"server-address"`
	ServerPort     uint           `mapstructure:"server-port"`
	RequestTimeout time.Duration  `mapstructure:"request-timeout"`
	Logger         *logrus.Logger `mapstructure:"-"`
}

func (cfg ClientConfig) BaseURL() string {
	return fmt.Sprintf("%s:%d", cfg.Address(), cfg.Port())
}

func (cfg ClientConfig) Address() string {
	if cfg.ServerAddress == "" {
		return DefaultServerAddress
	}

	return cfg.ServerAddress
}

func (cfg ClientConfig) Port() uint {
	if cfg.ServerPort == 0 {
		return DefaultServerPort
	}

	return cfg.ServerPort
}

func (cfg ClientConfig) Timeout() time.Duration {
	if cfg.RequestTimeout == 0 {
		return DefaultRequestTimeout
	}

	return cfg.RequestTimeout
}

func (cfg ClientConfig) Log() *logrus.Logger {
	if cfg.Logger == nil {
		return logrus.New()
	}

	return cfg.Logger
}
