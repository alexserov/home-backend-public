package main

import (
	modbusrelay "serov/home-backend-public/modbus/devices/relay"
)

func main() {
	relay:= modbusrelay.Create(243);
	
	relay.State()

	select {
	}
}
