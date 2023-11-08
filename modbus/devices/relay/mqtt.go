package relay

import (
	"fmt"

	mqttlib "github.com/eclipse/paho.mqtt.golang"
)

type mqttRelayMeta struct {
	topicToChannelMap map[string]byte
}

func (relay *relay)AttachToMqtt(mqttClient *mqttlib.Client) (err error){
	subscribeTopicMap := make(map[string]byte, 0)
	meta := mqttRelayMeta{topicToChannelMap: make(map[string]byte, 6)}

	for i:=0; i<6; i++ {
		topic := fmt.Sprintf("devices/%s/%d/k%d/on", relay.relayType, relay.slaveId, i+1)
		subscribeTopicMap[topic] = 0
		meta.topicToChannelMap[topic] = byte(i)
	}
	(*mqttClient).SubscribeMultiple(subscribeTopicMap, func(c mqttlib.Client, m mqttlib.Message) {
		relay.onMessage(m)
	})
	relay.meta["mqtt"] = meta


	relay.StateChanged().Add(func(sender Relay, args StateChangedArgs) {
		for i, v := range args.New.Outputs {
			topic := fmt.Sprintf("devices/%s/%d/k%d/on", relay.relayType, relay.slaveId, i+1)
			val := 1
			if(!v) {
				val = 0
			}
			(*mqttClient).Publish(topic, 1, true, val)
		}
	})
	return nil
}
func (relay *relay)OnMqttConnect() (err error){
	return nil
}
func (relay *relay)OnMqttConnectionLost() (error){
	return nil
}

func (relay *relay)onMessage(m mqttlib.Message){
	channel := (relay.meta["mqtt"]).(mqttRelayMeta).topicToChannelMap[m.Topic()]
	relay.Set(channel, string(m.Payload()) == "1")

	m.Ack()
}