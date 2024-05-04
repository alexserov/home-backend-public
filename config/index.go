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

type Config struct {
	Auth AuthConfig
}

var once sync.Once
var instance Config
var factories map[string]func()interface{} = make(map[string]func() interface{})

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