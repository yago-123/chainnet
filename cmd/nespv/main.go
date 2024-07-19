package main

import (
	"chainnet/cmd/nespv/cmd"

	"github.com/sirupsen/logrus"
)

var logger = logrus.New()

func main() {
	cmd.Execute(logger)
}
