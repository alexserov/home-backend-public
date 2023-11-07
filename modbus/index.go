package modbus

import (
	"serov/home-backend-public/config"
	"serov/home-backend-public/modbus/devices/relay"
)

func RegisterConfigFactory() {
	relay.RegisterConfigFactory()
}

func Initialize() {
	for _, device := range config.Get().Devices {
		relay.FromConfig(device.Id, *(device.Config).(*relay.RelayConfig))
	}
}