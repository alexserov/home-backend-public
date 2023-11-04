package manager

import (
	devicesRefreshable "serov/home-backend-public/modbus/devices/refreshable"
	"sync"
	"time"
)

var once sync.Once
var instance *manager

type manager struct {
	items []devicesRefreshable.Refreshable
	ticker *time.Ticker
	disposeChannel chan struct{}
	disposed bool
}

type Manager interface {
	Register(item devicesRefreshable.Refreshable)
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

func (manager *manager) Register(item devicesRefreshable.Refreshable) {
	manager.items = append(manager.items, item)
}