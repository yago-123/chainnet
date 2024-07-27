package config

import (
	"time"

	"github.com/sirupsen/logrus"
)

type Config struct {
	Logger         *logrus.Logger
	MiningInterval time.Duration
}

func NewConfig(logger *logrus.Logger, miningInterval time.Duration) *Config {
	return &Config{
		Logger:         logger,
		MiningInterval: miningInterval,
	}
}
