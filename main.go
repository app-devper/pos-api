package main

import (
	"pos/app"

	"github.com/sirupsen/logrus"
)

func main() {
	var server app.Routes
	if err := server.StartGin(); err != nil {
		logrus.WithError(err).Fatal("failed to start server")
	}
}
