package relay

import (
	"bytes"
	"encoding/binary"
	"serov/home-backend-public/modbus/devices/manager"
	devicesRefreshable "serov/home-backend-public/modbus/devices/refreshable"
	modbusQueue "serov/home-backend-public/modbus/queue"
	event "serov/home-backend-public/utils/event"

	"github.com/google/go-cmp/cmp"

	"github.com/grid-x/modbus"
)

var queue = modbusQueue.Instance()

type relay struct {
	slaveId byte
	state State
	stateChanged event.EventPrivate[Relay, StateChangedArgs]
}

type Relay interface {
	devicesRefreshable.Refreshable
	Set(index byte, value bool) error
	SetAll(values [6]bool) error
	State() State
	StateChanged() event.EventPublic[Relay, StateChangedArgs]
}

func Create(slaveId byte) Relay {
	result := &relay{slaveId: slaveId}
	result.stateChanged = event.Create[Relay, StateChangedArgs](result)

	result.initialize()

	return result
} 

type generic[V any] struct {
	relay *relay
}

func (generic generic[V]) invokeGeneric(action func(client modbus.Client) (V, error)) (V, error){
	cResults := make (chan V)
	defer close(cResults)

	cErr := make(chan error)
	defer close(cErr)

	queue.Enqueue(generic.relay.slaveId, func (client modbus.Client)  {
		r,e := action(client)
		cResults <- r
		cErr <- e
	})

	results := <-cResults
	err :=  <-cErr

	return results, err
}
func (relay *relay) invoke(action func(client modbus.Client) ([] byte, error)) ([] byte, error){
	return generic[[]byte]{relay}.invokeGeneric(action)
}

func (relay *relay) initialize() {
	val, err:= relay.invoke(func (cl modbus.Client) ([]byte, error)  {
		return cl.ReadHoldingRegisters(128, 1)
	})
	if err != nil {
		panic("has error")
	}

	if val[1] != relay.slaveId {
		panic("invalid slave id")
	}

	manager.Instance().Register(relay)
}

func (relay *relay) State() State {
	return relay.state;
}

func (relay *relay) Refresh() {
	oldState := relay.State()

	newState, err:= generic[State]{relay}.invokeGeneric(func (cl modbus.Client) (State, error)  {
		result := relay.State()

		outputs, coilsErr := cl.ReadCoils(0, 6)

		if coilsErr == nil {
			for i := 0; i<6; i++ {
				result.Outputs[i] = outputs[0] & (1 << byte(i)) > 0
			}
		}

		inputs, inputsErr := cl.ReadDiscreteInputs(0, 8)
		if inputsErr == nil {
			for i := 0; i<6; i++ {
				result.Inputs[i] = inputs[0] & (1 << byte(i)) > 0
			}
			result.Inputs[0] = inputs[0] & (1 << 7) > 0
		}

		counters, countersErr := cl.ReadInputRegisters(32, 8)
		uCounters := make([]uint16, 8)
		binary.Read(bytes.NewReader(counters), binary.BigEndian, uCounters)
		if countersErr == nil {
			for i, val := range uCounters[:5] {
				result.Counters[i+1] = val
			}
			result.Counters[0] = uCounters[7]
		}

		return result, nil
	})

	if err == nil {
		if !cmp.Equal(oldState, newState){
			relay.raiseStateChanged(StateChangedArgs{oldState, newState})
		}
		relay.state = newState
	}
}

func (relay *relay)raiseStateChanged(args StateChangedArgs) {
	relay.stateChanged.RaiseEvent(args)
}

func (relay *relay)StateChanged()event.EventPublic[Relay, StateChangedArgs]{
	return relay.stateChanged
}

func (relay *relay)Set(index byte, value bool) error { 
	_, err := generic[bool]{relay}.invokeGeneric(func(client modbus.Client) (bool, error) {
		if value {
			_, err := client.WriteSingleCoil(uint16(index), 1)
			return true, err
		} else {
			_, err := client.WriteSingleCoil(uint16(index), 0)
			return true, err
		}
	})
	return err
}

func (relay *relay)SetAll(values [6]bool) error { 
	_, err := generic[bool]{relay}.invokeGeneric(func(client modbus.Client) (bool, error) {
		result := byte(0)
		for i,v :=range values {
			if !v { continue; }
			result = result | 1 << i
		}
		_, err := client.WriteMultipleCoils(0, 6, []byte{result})
		return true, err
	})
	return err
}