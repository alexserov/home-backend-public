package queue

import (
	"sync"

	"github.com/grid-x/modbus"
)

type callback func(cl modbus.Client)

type queueAction struct {
	slaveId byte
	action callback
}

type queue struct {
	destroyed bool
	clientHandler modbus.ClientHandler
	client modbus.Client
	actionsSlow []queueAction
	actionsFast []queueAction
	processing bool
	processingMutex sync.Mutex
}
