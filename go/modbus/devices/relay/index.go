package relay

import (
	"bytes"
	"encoding/binary"
	"serov/home-backend-public/modbus/devices/manager"
	devicesRefreshable "serov/home-backend-public/modbus/devices/refreshable"
	modbusQueue "serov/home-backend-public/modbus/queue"

	"github.com/grid-x/modbus"
)

var queue = modbusQueue.Instance()

type relay struct {
	slaveId byte
	state state
}

type Relay interface {
	devicesRefreshable.Refreshable
	State() state
}

func Create(slaveId byte) Relay {
	result := &relay{slaveId: slaveId}
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

func (relay *relay) State() state {
	return relay.state;
}

func (relay *relay) Refresh() {
	val, err:= generic[state]{relay}.invokeGeneric(func (cl modbus.Client) (state, error)  {
		newState := relay.State()

		outputs, coilsErr := cl.ReadCoils(0, 6)

		if coilsErr == nil {
			for i := 0; i<6; i++ {
				newState.outputs[i] = outputs[0] & (1 << byte(i)) > 0
			}
		}

		inputs, inputsErr := cl.ReadDiscreteInputs(0, 8)
		if inputsErr == nil {
			for i := 0; i<6; i++ {
				newState.inputs[i] = inputs[0] & (1 << byte(i)) > 0
			}
			newState.inputs[0] = inputs[0] & (1 << 7) > 0
		}

		counters, countersErr := cl.ReadInputRegisters(32, 8)
		uCounters := make([]uint16, 8)
		binary.Read(bytes.NewReader(counters), binary.BigEndian, uCounters)
		if countersErr == nil {
			for i, val := range counters[:5] {
				newState.inputs[i+1] = val>0
			}
			newState.inputs[0] = counters[7] > 0
		}

		return newState, nil
	})

	if err == nil {
		relay.state = val
	}
}