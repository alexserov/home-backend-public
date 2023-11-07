package relay

import (
	"serov/home-backend-public/config"
	"serov/home-backend-public/modbus/devices/manager"
)

func RegisterConfigFactory() {
	config.Register("Modbus6chRelay", func() interface{} { return new(RelayConfig)})
}

type RelayConfigEventHandler struct {
	Action string
	Target byte
	Switches []byte
}
type RelayConfigEvent struct {
	Channel byte
	EventName string
	Handlers []RelayConfigEventHandler
}
type RelayConfig struct {
	Events []RelayConfigEvent
}

func FromConfig(id byte, config RelayConfig) Relay {
	result := Create(id)

	result.StateChanged().Add(func(sender Relay, args StateChangedArgs) {
		for _, rule := range config.Events {
			changed := false
			switch rule.EventName {
				case "Click": {
					changed = args.Old.Clicks[rule.Channel] != args.New.Clicks[rule.Channel]
				}
				default: {
					panic("unsupported action")
				}
			}

			if !changed {
				continue
			}

			for _, handler := range rule.Handlers {
				target := manager.Instance().Get(byte(handler.Target)).(Relay)
				current := target.State().Outputs
				if handler.Action == "on" {
					for _, switchIndex := range handler.Switches {
						current[switchIndex] = true
					}
				}

				if handler.Action == "off" {
					for _, switchIndex := range handler.Switches {
						current[switchIndex] = false
					}
				}

				if handler.Action == "toggle" {
					for _, switchIndex := range handler.Switches {
						current[switchIndex] = !current[switchIndex]
					}
				}


				target.SetAll(current)
			}

		}
	})

	return result
}