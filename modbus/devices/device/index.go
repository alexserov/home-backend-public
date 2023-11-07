package device

import (
	"serov/home-backend-public/modbus/devices/refreshable"
)

type Device interface {
	refreshable.Refreshable
	Id() byte
}