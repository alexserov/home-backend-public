package queue

import (
	"sync"

	"github.com/grid-x/modbus"
	"go.uber.org/zap"
)

var instance *queue
var once sync.Once

type Queue interface {
	Enqueue(fast bool, slaveId byte, item callback) Queue
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

func (q *queue) enqueueAsync(fast bool, slaveId byte, item callback) {
	q.assertNotDestroyed()

	q.appendLocked(fast, slaveId, item)
	go q.ProcessItems()
}
func (q *queue) Enqueue(fast bool, slaveId byte, item callback) Queue {
	q.assertNotDestroyed()

	go q.enqueueAsync(fast, slaveId, item)

	return q
}

func (q *queue) appendLocked(fast bool, slaveId byte, item callback) {
	q.mutateActionsMutex.Lock()
	defer q.mutateActionsMutex.Unlock()

	if fast {
		q.actionsFast = append(q.actionsFast, queueAction{slaveId, item})
	} else {
		q.actionsSlow = append(q.actionsSlow, queueAction{slaveId, item})
	}
}

func (q *queue) processQueueActions(actions *[]queueAction) bool {
	if len(*actions) > 0 {
		if !q.mutateActionsMutex.TryLock() {
			zap.L().Debug("already locked")
			return false
		}
		meta := (*actions)[0]
		*actions = (*actions)[1:]
		q.mutateActionsMutex.Unlock()

		q.clientHandler.SetSlave(meta.slaveId)
		meta.action(q.client)
		return true
	}
	return false
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
	defer func(_q *queue) {
		_q.processing = false
	}(q)

	for len(q.actionsFast) > 0 || len(q.actionsSlow) > 0 {
		if q.processQueueActions(&q.actionsFast) {
			continue
		}
		q.processQueueActions(&q.actionsSlow)
	}

	return q
}
