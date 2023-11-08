package device

import (
	mqttlib "github.com/eclipse/paho.mqtt.golang"
)

type Device interface {
	AttachToMqtt(manager *mqttlib.Client) (err error)
	OnMqttConnect() (err error)
	OnMqttConnectionLost() (error)
}