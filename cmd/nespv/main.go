package main

import (
	"github.com/sirupsen/logrus"
	"github.com/yago-123/chainnet/cmd/nespv/cmd"
)

func main() {
	cmd.Execute(logrus.New())
}
