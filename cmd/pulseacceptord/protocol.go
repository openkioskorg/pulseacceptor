package main

import (
	"encoding/json"
	"log"

	"github.com/eclipse/paho.golang/paho"
)

type PulseacceptordEvent struct {
	Amount uint64 `json:"amount"`
}

type PulseacceptordCommand struct {
	Accept bool `json:"accept"`
}

func commandHandler(msg *paho.Publish) {
	var cmd PulseacceptordCommand
	if err := json.Unmarshal(msg.Payload, &cmd); err != nil {
		log.Printf("Command could not be parsed (%s): %s", msg.Payload, err)
	}
	if cmd.Accept {
		log.Println("Received start command, counting coin input")
		accept = true
	} else {
		log.Println("Received stop command, ignoring coin input")
		accept = false
	}
}
