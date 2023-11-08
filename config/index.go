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
	Login string
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

var once sync.Once
var instance Config
var factories map[string]func()interface{} = make(map[string]func() interface{})

func (c *DeviceConfig) UnmarshalJSON(data []byte) (err error) {
	var deviceFields map[string]*json.RawMessage
	json.Unmarshal(data, &deviceFields)

	json.Unmarshal(*deviceFields["Id"], &c.Id)
	json.Unmarshal(*deviceFields["Type"], &c.Type)

	factory := factories[c.Type]
	c.Config = factory()
	json.Unmarshal(*deviceFields["Config"], c.Config)

	return nil
}

func readConfig() {
	file, error := os.ReadFile("./private/config.json")

	if error != nil {
		panic(error.Error())
	}

	json.Unmarshal(file, &instance)
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