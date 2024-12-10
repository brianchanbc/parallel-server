package main

import (
	"encoding/json"
	"os"
	"parserver/server"
	"strconv"
)

func main() {
	// Get arguments
	args := os.Args
	// Set up configuration requirements
	config := server.Config{}
	// Set default as sequential version
	config.Mode = "s"
	config.ConsumersCount = 1
	config.Encoder = json.NewEncoder(os.Stdout)
	config.Decoder = json.NewDecoder(os.Stdin)
	if len(args) > 1 {
		numConsumers, err := strconv.Atoi(args[1])
		if err == nil && numConsumers > 1 {
			// Change to parallel version
			config.Mode = "p"
			config.ConsumersCount = numConsumers
		}
	}
	server.Run(config)
}
