package relay

import (
	modbusDevice "serov/home-backend-public/modbus/devices/device"
	mqttDevice "serov/home-backend-public/mqtt/device"
	event "serov/home-backend-public/utils/event"
)

type relay struct {
	slaveId byte
	relayType string
	state State
	meta map[string]interface{}
	stateChanged event.EventPrivate[Relay, StateChangedArgs]
}

type Relay interface {
	modbusDevice.Device
	mqttDevice.Device
	Set(index byte, value bool) error
	SetAll(values [6]bool) error
	State() State
	StateChanged() event.EventPublic[Relay, StateChangedArgs]
}

func Create(slaveId byte, relayType string) Relay {
	result := &relay{slaveId: slaveId, relayType: relayType, meta: make(map[string]interface{})}
	result.stateChanged = event.Create[Relay, StateChangedArgs](result)

	result.modbusInitialize()

	return result
} 

func (relay *relay) State() State {
	return relay.state;
}

func (relay *relay)raiseStateChanged(args StateChangedArgs) {
	relay.stateChanged.RaiseEvent(args)
}

func (relay *relay)StateChanged()event.EventPublic[Relay, StateChangedArgs]{
	return relay.stateChanged
}

func (relay *relay) Id() byte {
	return relay.slaveId
}