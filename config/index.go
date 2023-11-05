package config

import (
	"encoding/json"
	"os"
	"sync"
)

type Config struct {
	mqtt struct {
		login string
		password string
	}
}

var once sync.Once
var instance Config

func readConfig() {
	file, error := os.ReadFile("../private/config.json")

	if error != nil {
		panic(error.Error())
	}


	json.Unmarshal(file, &instance)
}

func Get() Config {
	once.Do(func() {
		readConfig()
	})

	return instance
}