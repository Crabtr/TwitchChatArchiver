package main

import (
	"encoding/json"
	"io/ioutil"
)

type config struct {
	Nick     string   `json:"nick"`
	OAuth    string   `json:"oauth"`
	Channels []string `json:"channels"`
}

func loadConfig() (config, error) {
	var config config

	// Open config for reading
	configFile, err := ioutil.ReadFile("config.json")
	if err != nil {
		return config, err
	}

	// Unmarshal config
	err = json.Unmarshal(configFile, &config)
	if err != nil {
		return config, err
	}

	return config, nil
}
