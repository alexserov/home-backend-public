package mqtt

import (
	"fmt"
	"sync"

	"serov/home-backend-public/config"
	"serov/home-backend-public/mqtt/device"

	mqttlib "github.com/eclipse/paho.mqtt.golang"
)

var instance manager
var once sync.Once

type manager struct {
	devices []*device.Device
	client *mqttlib.Client
}

type Manager interface {
	Register(device device.Device) (err error)
}	

func Instance() Manager {
	once.Do((func() {
		instance.devices = make([]*device.Device, 0)
		instance.initialize()
	}))
	return &instance
}

func (manager *manager)initialize() {
	var broker = "localhost"
    var port = 1883
    opts := mqttlib.NewClientOptions()
    opts.AddBroker(fmt.Sprintf("tcp://%s:%d", broker, port))
    opts.SetClientID("go_mqtt_client")
    opts.SetUsername(config.Get().Auth.Mqtt.Login)
    opts.SetPassword(config.Get().Auth.Mqtt.Password)
 
	opts.OnConnect = func(c mqttlib.Client) {
		for _, dev := range manager.devices {
			(*dev).OnMqttConnect()
		}
	}
    opts.OnConnectionLost = func(c mqttlib.Client, err error) {
		for _, dev := range manager.devices {
			(*dev).OnMqttConnectionLost()
		}
	}
    client := mqttlib.NewClient(opts)
	manager.client = &client;
    if token := client.Connect(); token.Wait() && token.Error() != nil {
        panic(token.Error())
	}
}

func (manager *manager) Register(device device.Device) (err error) {
	manager.devices = append(manager.devices, &device)

	device.AttachToMqtt(manager.client)
	return nil
}