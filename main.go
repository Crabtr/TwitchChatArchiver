package main

import (
	"log"
	"os"
	"sync"
	"time"
)

func main() {
	// Load the config
	config, err := loadConfig()
	if err != nil {
		panic(err)
	}

	// Verify ./logs/ exists
	if _, err := os.Stat("logs/"); os.IsNotExist(err) {
		os.Mkdir("logs/", 0775)

		log.Printf("Created directory: logs/\n")
	}

	// Verify ./logs/channel exists for every channel
	for idx := range config.Channels {
		if _, err := os.Stat("logs/" + config.Channels[idx][1:]); os.IsNotExist(err) {
			os.Mkdir("logs/"+config.Channels[idx][1:], 0755)

			log.Printf("Created directory: logs/%s\n", config.Channels[idx][1:])
		}
	}

	// Create a self-contained connection for each channel
	var wg sync.WaitGroup
	for idx := range config.Channels {
		wg.Add(1)
		go connect(wg, config.Nick, config.OAuth, config.Channels[idx])

		time.Sleep(1 * time.Second)
	}

	wg.Wait()
}
