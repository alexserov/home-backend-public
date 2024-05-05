package queue

import (
	"sync"

	"github.com/grid-x/modbus"
	"go.uber.org/zap"
)

var instance *queue
var once sync.Once

type Queue interface {
	Enqueue(slaveId byte, item callback) Queue
	Destroy() Queue
}

func Instance() Queue {
	once.Do(func() {
		instance = &queue{}
		instance.initialize()
	})
	return instance
}

func rtuInitialize() *modbus.RTUClientHandler {
	var handler = modbus.NewRTUClientHandler("/dev/ttyACM0")
	handler.BaudRate = 9600
	handler.DataBits = 8
	handler.StopBits = 2
	handler.Parity = "N"

	return handler
}
func tcpInitialize() *modbus.TCPClientHandler {
	var handler = modbus.NewTCPClientHandler("localhost:5020")

	return handler
}

func (q *queue) initialize() Queue {
	q.clientHandler = rtuInitialize()

	q.clientHandler.Connect()

	q.client = modbus.NewClient(q.clientHandler)

	return q
}

func (q *queue) assertNotDestroyed() {
	if q.destroyed {
		panic("Queue is already destroyed")
	}
}

func (q *queue) Destroy() Queue {
	q.assertNotDestroyed()

	q.clientHandler.Close()
	q.destroyed = true
	return q
}

func (q *queue) Enqueue(slaveId byte, item callback) Queue {
	q.assertNotDestroyed()

	q.actions = append(q.actions, queueAction{slaveId, item})
	go q.ProcessItems()
	return q
}

func (q *queue) ProcessItems() Queue {
	q.assertNotDestroyed()

	if q.processing {
		return q
	}
	
	if !q.processingMutex.TryLock() {
		zap.L().Debug("queue already locked")
		return q
	}
	if q.processing {
		return q
	}
	defer q.processingMutex.Unlock()
	q.processing = true
	for len(q.actions) > 0 {
		meta := q.actions[0]
		q.actions = q.actions[1:]

		q.clientHandler.SetSlave(meta.slaveId)
		meta.action(q.client)
	}
	q.processing = false

	return q

}
