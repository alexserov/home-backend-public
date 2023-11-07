package manager

import (
	"serov/home-backend-public/modbus/devices/device"
	"sync"
	"time"
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
	manager.ticker = time.NewTicker(200 * time.Millisecond)
	manager.disposeChannel = make (chan struct{})
	go func () {
		for {
			select {
			case <-manager.ticker.C:
				manager.processActions()
			case <-manager.disposeChannel:
				manager.ticker.Stop()
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

}

func (manager *manager) Register(item device.Device) {
	manager.items = append(manager.items, item)
	manager.itemIdToItemMap[item.Id()] = item
}

func (manager *manager) Get(id byte) device.Device {
	return manager.itemIdToItemMap[id];
}

