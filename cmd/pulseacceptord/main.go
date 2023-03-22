package main

import (
	"context"
	"log"

	//pa "gitlab.com/openkiosk/pulseacceptor"
	"periph.io/x/host/v3"
)

var accept bool // true when accepting, false on idle

func main() {
	conf := parseConfig()
	if _, err := host.Init(); err != nil {
		log.Fatal("Failed to load host drivers: ", err)
	}

	/*	pulseDevice, err := pa.Init(conf.Device)
		if err != nil {
			log.Fatal("Failed to initialize pulse device: ", err)
		}
	*/

	broker, err := newBroker(conf.Mqtt)
	if err != nil {
		log.Fatal("Failed to connect to MQTT broker: ", err)
	}

	pulseChan := make(chan uint64)
	//go pulseDevice.CountWithHandler(pulseChan)
	go bsvalues(pulseChan)

	accept = true
	for {
		select {
		case p := <-pulseChan:
			amount := conf.Values[p]
			if accept {
				log.Printf("Received %d cents.\n", amount)
				if err := broker.publishAmount(context.Background(), amount); err != nil {
					log.Println("Failed to publish event: ", err)
				}
			}
		}
	}
}
