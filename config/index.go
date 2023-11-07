package config

import (
	"encoding/json"
	"os"
	"sync"
)

type AuthConfig struct {
	Mqtt Mqtt
}
type Mqtt struct {
	Login string `json:"asd"`
	Password string
}

type DeviceConfig struct {
	Id byte
	Type string
	Config interface{}
}

type Config struct {
	Auth AuthConfig
	Devices []DeviceConfig
}

type DeviceConfigRaw struct {
	Id byte
	Type string
	Config json.RawMessage
}

type ConfigRaw struct {
	Auth AuthConfig
	Devices []DeviceConfigRaw
}

var once sync.Once
var instance Config
var factories map[string]func()interface{} = make(map[string]func() interface{})

func readConfig() {
	file, error := os.ReadFile("./private/config.json")

	if error != nil {
		panic(error.Error())
	}

	var rawInstance ConfigRaw
	instance.Auth = rawInstance.Auth

	json.Unmarshal(file, &rawInstance)
	for _,deviceRaw := range rawInstance.Devices {
		factory := factories[deviceRaw.Type]
		deviceConfig := factory()
		json.Unmarshal(deviceRaw.Config, deviceConfig)
		
		device := DeviceConfig{deviceRaw.Id, deviceRaw.Type, deviceConfig}

		instance.Devices = append(instance.Devices, device)
	}
}

func Register(typeKey string, factory func()interface{}) {
	factories[typeKey] = factory
}

func Get() Config {
	once.Do(func() {
		readConfig()
	})

	return instance
}