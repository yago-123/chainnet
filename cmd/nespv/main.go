package main

import (
	"chainnet/cmd/nespv/cmd"
	"github.com/sirupsen/logrus"
)

var logger = logrus.New()

func main() {
	cmd.Execute(logger)

	// initialize network for sending transactions to the miners
	// p2p.NewP2PNode(context.Background(), &config.Config{})
}
