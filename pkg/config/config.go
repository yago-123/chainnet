package config

import "github.com/sirupsen/logrus"

type Config struct {
	Logger *logrus.Logger
	// todo() split POW into separate config
	DifficultyPoW uint
	MaxNoncePoW   uint
}

func NewConfig(logger *logrus.Logger, difficultyPoW uint, maxNoncePow uint) *Config {
	return &Config{
		Logger:        logger,
		DifficultyPoW: difficultyPoW,
		MaxNoncePoW:   maxNoncePow,
	}
}
