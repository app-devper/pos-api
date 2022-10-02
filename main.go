package main

import (
	"pos/app"
)

func main() {
	var server app.Routes
	server.StartGin()
}
