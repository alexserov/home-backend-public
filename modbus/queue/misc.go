package queue

import (
	"github.com/grid-x/modbus"
)

type callback func(cl modbus.Client)

type queueAction struct {
	slaveId byte
	action callback
}

type queue struct {
	destroyed bool
	clientHandler *modbus.RTUClientHandler
	client modbus.Client
	actions []queueAction
	processing bool
}
