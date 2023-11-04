package main

import (
	modbusrelay "serov/home-backend-public/modbus/devices/relay"
)

func main() {
	relay:= modbusrelay.Create(243);
	
	relay.StateChanged().Add(func(sender modbusrelay.Relay, args modbusrelay.StateChangedArgs) {
		if args.Old.Counters[0] < args.New.Counters[0] {
			println("changed")
			if args.New.Outputs[0] {
				sender.SetAll([6]bool{})
			} else {
				sender.SetAll([6]bool{true, true, true})
			}
		}
	})

	select {
	}
}
