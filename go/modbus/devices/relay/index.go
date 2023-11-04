package relay

import (
	modbusQueue "serov/home-backend-public/modbus/queue"
	devicesRefreshable "serov/home-backend-public/modbus/devices/refreshable"

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

func (relay *relay) invoke(action func(client modbus.Client) ([]byte, error)) ([]byte, error){
	cResults := make (chan []byte)
	defer close(cResults)

	cErr := make(chan error)
	defer close(cErr)

	queue.Enqueue(relay.slaveId, func (client modbus.Client)  {
		r,e := action(client)
		cResults <- r
		cErr <- e
	})

	results := <-cResults
	err :=  <-cErr

	return results, err
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
}

func (relay *relay) State() state {
	return relay.state;
}

func (relay *relay) Refresh() {
	
}