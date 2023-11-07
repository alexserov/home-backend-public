package mqtt

import (
	"fmt"
	"log"
	"sync"

	mqttlib "github.com/eclipse/paho.mqtt.golang"
)

var instance listener
var once sync.Once

type listener struct {
}

type Listener interface {

}	

func Create() Listener {
	once.Do((func() {
		instance.initialize()
	}))
	return instance
}

func (listener *listener)initialize() {
	var broker = "localhost"
    var port = 1883
    opts := mqttlib.NewClientOptions()
    opts.AddBroker(fmt.Sprintf("tcp://%s:%d", broker, port))
    opts.SetClientID("go_mqtt_client")
    opts.SetUsername("emqx")
    opts.SetPassword("public")
    opts.SetDefaultPublishHandler(func(c mqttlib.Client, m mqttlib.Message) {
		log.Default().Println("pub")
	})
    opts.OnConnect = func(c mqttlib.Client) {

	}
    opts.OnConnectionLost = func(c mqttlib.Client, err error) {

	}
    client := mqttlib.NewClient(opts)
    if token := client.Connect(); token.Wait() && token.Error() != nil {
        panic(token.Error())
	}
	token:= client.Subscribe("test/+", 1, func(c mqttlib.Client, m mqttlib.Message) {
		fmt.Printf("got message %s", string(m.Payload()))
	});
	token.Wait();

}