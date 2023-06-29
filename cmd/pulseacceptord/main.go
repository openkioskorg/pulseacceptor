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
	"os"
	"os/signal"
	"syscall"

	pa "gitlab.com/openkiosk/pulseacceptor"
	"periph.io/x/conn/v3/gpio"
	"periph.io/x/conn/v3/gpio/gpioreg"
	"periph.io/x/host/v3"
)

var accept bool // true when accepting, false on idle

var enablePin gpio.PinIO

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Usage: ./pulseacceptord config.yaml")
	}
	conf = parseConfig(os.Args[1])

	if _, err := host.Init(); err != nil {
		log.Fatal("Failed to load host drivers: ", err)
	}

	pulseDevice, err := pa.Init(conf.Device)
	if err != nil {
		log.Fatal("Failed to initialize pulse device: ", err)
	}

	if conf.EnablePinControl {
		enablePin = gpioreg.ByName(conf.EnablePin)
	}

	broker, err := newBroker(conf.Mqtt)
	if err != nil {
		log.Fatal("Failed to connect to MQTT broker: ", err)
	}

	pulseChan := make(chan int64)
	go pulseDevice.CountWithHandler(pulseChan)

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c

		// At exit disable the pulse device if possible
		if conf.EnablePinControl {
			toWrite := gpio.High
			if conf.EnabledWhenHigh {
				toWrite = gpio.Low
			}
			if err := enablePin.Out(toWrite); err != nil {
				log.Println("Failed to disable pulse device")
			}
		}
		os.Exit(1)
	}()

	for {
		select {
		case p := <-pulseChan:
			amount := conf.Values[p]
			if accept {
				log.Printf("Received amount: %d\n", amount)
				if err := broker.publishAmount(context.Background(), amount); err != nil {
					log.Println("Failed to publish event: ", err)
				}
			}
		}
	}
}
