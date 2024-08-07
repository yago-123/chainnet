package config

import (
	"time"

	"github.com/sirupsen/logrus"
)

type Config struct {
	Logger         *logrus.Logger
	MiningInterval time.Duration
	P2PEnabled     bool
	// P2PMinNumConn minimum number of connections to maintain by connection manager
	P2PMinNumConn uint
	// P2PMaxNumConn maximum number of connections to maintain by connection manager
	P2PMaxNumConn uint
}

func NewConfig(
	logger *logrus.Logger,
	miningInterval time.Duration,
	p2pEnabled bool,
	p2pMinNumConn uint,
	p2pMaxNumConn uint,
) *Config {
	return &Config{
		Logger:         logger,
		MiningInterval: miningInterval,
		P2PEnabled:     p2pEnabled,
		P2PMinNumConn:  p2pMinNumConn,
		P2PMaxNumConn:  p2pMaxNumConn,
	}
}
