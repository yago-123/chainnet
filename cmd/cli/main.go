package main

import (
	"chainnet/cmd/cli/cmd"

	"github.com/sirupsen/logrus"
)

var logger = logrus.New()

func main() {
	cmd.Execute(logger)
}
