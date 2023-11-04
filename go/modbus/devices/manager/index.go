package manager

import (
	"sync"
)

var once sync.Once
var instance *manager

type manager struct {

}

type Manager interface {

}

func Instance() Manager {
	once.Do(func() {
		instance = &manager{}
	})

	return instance
}

