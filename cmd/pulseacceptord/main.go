/* Daemon for counting pulses from money acceptors.
   Copyright (C) 2023  Digilol OÜ

   This program is free software: you can redistribute it and/or modify
   it under the terms of the GNU Affero General Public License as
   published by the Free Software Foundation, either version 3 of the
   License, or (at your option) any later version.

   This program is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU Affero General Public License for more details.

   You should have received a copy of the GNU Affero General Public License
   along with this program.  If not, see <https://www.gnu.org/licenses/>. */

package main

import (
	"context"
	"log"

	pa "gitlab.com/openkiosk/pulseacceptor"
	"periph.io/x/host/v3"
)

var accept bool // true when accepting, false on idle

func main() {
	conf := parseConfig()
	if _, err := host.Init(); err != nil {
		log.Fatal("Failed to load host drivers: ", err)
	}

	pulseDevice, err := pa.Init(conf.Device)
	if err != nil {
		log.Fatal("Failed to initialize pulse device: ", err)
	}

	broker, err := newBroker(conf.Mqtt)
	if err != nil {
		log.Fatal("Failed to connect to MQTT broker: ", err)
	}

	pulseChan := make(chan int64)
	go pulseDevice.CountWithHandler(pulseChan)

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
