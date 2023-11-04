package main

import (
	modbusqueue "serov/home-backend-public/modbus/queue"
	modbusrelay "serov/home-backend-public/modbus/devices/relay"
)

func main() {
	queue:= modbusqueue.Instance()

	relay:= modbusrelay.Create(243);
	
	relay.State()
	
	defer queue.Destroy()
}
