package config

import "github.com/sirupsen/logrus"

type Config struct {
	Logger        *logrus.Logger
	DifficultyPoW uint
	MaxNoncePoW   uint
	BaseURL       string
}

func NewConfig(logger *logrus.Logger, difficultyPoW uint, maxNoncePoW uint, baseURL string) *Config {
	return &Config{
		Logger:        logger,
		DifficultyPoW: difficultyPoW,
		MaxNoncePoW:   maxNoncePoW,
		BaseURL:       baseURL,
	}
}
