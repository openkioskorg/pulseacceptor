package main

import (
	"encoding/json"
	"log"

	pa "gitlab.com/openkiosk/pulseacceptor"
	"periph.io/x/host/v3"
)

type pulseAcceptorEvent struct {
	Amount uint64 `json:"amount"`
}

func main() {
	conf := parseConfig()
	if _, err := host.Init(); err != nil {
		log.Fatal("Failed to load host drivers: ", err)
	}

	pulseDevice, err := pa.Init(conf.Device)
	if err != nil {
		log.Fatal("Failed to initialize pulse device: ", err)
	}

	queue := initQueue(conf.Redis)

	pulseChan := make(chan uint64)
	go pulseDevice.CountWithHandler(pulseChan)

	for {
		select {
		case p := <-pulseChan:
			log.Printf("Received %d pulses.\n", p)
			j, _ := json.Marshal(pulseAcceptorEvent{Amount: conf.Values[p]})
			if err := queue.PublishBytes(j); err != nil {
				log.Println("Failed to publish event: ", err)
			}
		}
	}
}
