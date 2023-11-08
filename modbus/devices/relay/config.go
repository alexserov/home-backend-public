package relay

import (
	"errors"
	"serov/home-backend-public/config"
	"serov/home-backend-public/modbus/devices/manager"
	"serov/home-backend-public/mqtt"
)

func RegisterConfigFactory() {
	config.Register("modbus6chrelay", func() interface{} { return new(RelayConfig)})
}

type RelayConfigEventHandler struct {
	Action RelayEventHanlderAction
	Target byte
	Switches []byte
}
type RelayConfigEvent struct {
	Channel byte
	EventName RelayEventName
	Handlers []RelayConfigEventHandler
}
type RelayConfig struct {
	Events []RelayConfigEvent
}

const (
	On RelayEventHanlderAction = iota + 1
	Toggle
	Off
)
type RelayEventHanlderAction uint8

func (a *RelayEventHanlderAction) UnmarshalJSON(data []byte) (err error) {
	switch string(data) {
	case `"on"`: *a = On
	case `"off"`: *a = Off
	case `"toggle"`: *a = Toggle
	default: return errors.New("invalid Events.Handlers.Action value")
	}
	return nil
}

const (
	Click RelayEventName = iota + 1
	DoubleClick
	LongClick
)

type RelayEventName uint8

func (n *RelayEventName) UnmarshalJSON(data []byte) (err error) {
	switch string(data) {
		case `"click"`: *n = Click
		case `"doubleclick"`: *n = DoubleClick
		case `"longclick"`: *n = LongClick
		default: return errors.New("invalid Events.EventName value")
		}
		return nil
}

func FromConfig(id byte, relayType string, config RelayConfig) Relay {
	result := Create(id, relayType)

	result.StateChanged().Add(func(sender Relay, args StateChangedArgs) {
		for _, rule := range config.Events {
			changed := false
			switch rule.EventName {
				case Click: {
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
				if handler.Action == On {
					for _, switchIndex := range handler.Switches {
						current[switchIndex] = true
					}
				}

				if handler.Action == Off {
					for _, switchIndex := range handler.Switches {
						current[switchIndex] = false
					}
				}

				if handler.Action == Toggle {
					for _, switchIndex := range handler.Switches {
						current[switchIndex] = !current[switchIndex]
					}
				}


				target.SetAll(current)
			}

		}
	})

	mqtt.Instance().Register(result);
	
	return result
}