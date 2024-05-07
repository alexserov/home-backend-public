package manager

import (
	"serov/home-backend-public/modbus/devices/device"
	"sync"
	"time"

	"go.uber.org/zap"
)

var once sync.Once
var instance *manager

type manager struct {
	items []device.Device
	itemIdToItemMap map[byte]device.Device
	ticker *time.Ticker
	disposeChannel chan struct{}
	disposed bool
}

type Manager interface {
	Get(id byte)device.Device
	Register(item device.Device)
	Dispose()
}

func Instance() Manager {
	once.Do(func() {
		instance = &manager{}
		instance.initialize()
	})

	return instance
}

func (manager *manager)initialize() {
	manager.itemIdToItemMap = make(map[byte]device.Device)
	manager.ticker = time.NewTicker(400 * time.Millisecond)
	manager.disposeChannel = make (chan struct{})
	var processed uint64 = 0 
	go func () {
		for {
			select {
			case <-manager.ticker.C:
				manager.processActions()
				if processed % 1000 == 0 {
					zap.L().Debug("Manager processed", zap.Uint64("count", processed))
				}
				processed++
			case <-manager.disposeChannel:
				manager.ticker.Stop()
				zap.L().Debug("!!!GOT MANAGER STOP COMMAND!!!")
				manager.disposed = true
			}
		}
	}()
}

func (manager *manager)processActions() {
	if manager.disposed {
		return
	}

	for _, current := range manager.items {
		(current).Refresh()
	}
}

func (manager *manager)Dispose() {
	manager.disposeChannel <- struct{}{}
	close(manager.disposeChannel)

}

func (manager *manager) Register(item device.Device) {
	manager.items = append(manager.items, item)
	manager.itemIdToItemMap[item.Id()] = item
}

func (manager *manager) Get(id byte) device.Device {
	return manager.itemIdToItemMap[id];
}

