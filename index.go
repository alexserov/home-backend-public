package main

import (
	"serov/home-backend-public/config"
	"serov/home-backend-public/modbus"
	// "serov/home-backend-public/mqtt"
)

// modbusrelay "serov/home-backend-public/modbus/devices/relay"

func main() {

	modbus.RegisterConfigFactory()
	config.Get()

	modbus.Initialize()
	// mqtt.Create()

	select {
	}
}
