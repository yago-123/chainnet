package main

import (
	"chainnet/cmd/wallet/cmd"

	"github.com/sirupsen/logrus"
)

var logger = logrus.New()

func main() {
	cmd.Execute(logger)
}
