package queue

import (
	"time"
	"sync"
	"github.com/grid-x/modbus"
)

var instance *queue
var once sync.Once

type Queue interface {
	Enqueue(slaveId byte, item callback) Queue
	Destroy() Queue
}

func Instance() Queue {
	once.Do(func ()  {
		instance = &queue{}
		instance.initialize()
	})
	return instance
}

func (q *queue) initialize() Queue {
	q.clientHandler = modbus.NewRTUClientHandler("/dev/ttyACM0")
	q.clientHandler.BaudRate = 9600
	q.clientHandler.DataBits = 8
	q.clientHandler.StopBits = 2
	q.clientHandler.Parity = "N"
	q.clientHandler.Timeout = time.Millisecond * 1000

	q.clientHandler.Connect()

	q.client = modbus.NewClient(q.clientHandler);

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

	if(q.processing) {
		return q
	}
	q.processing = true
	for len(q.actions)	> 0 {
		meta := q.actions[0]

		q.clientHandler.SetSlave(meta.slaveId)
		meta.action(q.client)
		q.actions = q.actions[1:]
	}
	q.processing = false

	return q

}