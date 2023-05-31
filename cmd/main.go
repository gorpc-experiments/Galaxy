package main

import (
	"github.com/gorpc-experiments/ServiceCore"
	"github.com/gorpc-experiments/galaxy/src/service"
)

func main() {
	ServiceCore.SetupLogging()

	galaxy := service.NewGalaxy()

	ServiceCore.PublishMicroService(galaxy, false)
}