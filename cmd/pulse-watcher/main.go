/* Program for debugging pulse counting.
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
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"periph.io/x/conn/v3/gpio"
	"periph.io/x/conn/v3/gpio/gpioreg"
	"periph.io/x/conn/v3/gpio/gpioutil"
	"periph.io/x/conn/v3/physic"
	"periph.io/x/host/v3"
)

var (
	green  = "\033[32m"
	yellow = "\033[33m"
	reset  = "\033[0m"
)

type args struct {
	pin              string
	enablePin        string
	enablePinControl bool
	enabledWhenHigh  bool
	freq             int
	timeout          time.Duration
}

var conf args

func init() {
	flag.StringVar(&conf.pin, "pin", "1", "Pulse input pin")
	flag.BoolVar(&conf.enablePinControl, "enablepincontrol", false, "true or false")
	flag.BoolVar(&conf.enabledWhenHigh, "enabledwhenhigh", false, "State is enabled when high is outputted")
	flag.StringVar(&conf.enablePin, "enablepin", "2", "Pulse enable pin")
	flag.IntVar(&conf.freq, "freq", 10, "Poll frequency in hertz")
	flag.DurationVar(&conf.timeout, "timeout", 110*time.Millisecond,
		"Timeout value for ignoring long pauses between different coin/bill inputs")
	flag.Parse()
}

var pEnable gpio.PinIO

func main() {
	if _, err := host.Init(); err != nil {
		log.Fatal(err)
	}

	p := gpioreg.ByName(conf.pin)
	if p == nil {
		log.Fatal("Unknown pin")
	}

	if conf.enablePinControl {
		pEnable = gpioreg.ByName(conf.enablePin)
		if p == nil {
			log.Fatal("Unknown enable pin")
		}
		toWrite := gpio.Low
		if conf.enabledWhenHigh {
			toWrite = gpio.High
		}
		if err := pEnable.Out(toWrite); err != nil {
			log.Fatal(err)
		}
	}

	p = gpioutil.PollEdge(p, physic.Frequency(conf.freq)*physic.Hertz)
	if err := p.In(gpio.PullUp, gpio.BothEdges); err != nil {
		log.Fatal(err)
	}

	var (
		totalPulsesLength, totalPausesLength                 time.Duration
		pulsesThisTime, totalPulsesNumber, totalPausesNumber int64
	)

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c

		if conf.enablePinControl {
			toWrite := gpio.High
			if conf.enabledWhenHigh {
				toWrite = gpio.Low
			}
			if err := pEnable.Out(toWrite); err != nil {
				log.Fatal(err, "Failed to disable pulse device")
			}
		}
		fmt.Printf("...\n")
		if totalPulsesNumber == 0 || totalPausesNumber == 0 {
			os.Exit(1)
		}
		fmt.Printf("Pulse avg: %s\n", time.Duration(int64(totalPulsesLength)/totalPulsesNumber).String())
		fmt.Printf("Pause avg: %s\n", time.Duration(int64(totalPausesLength)/totalPausesNumber).String())
		os.Exit(1)
	}()

	eventStart := time.Now()
	duringPulse := false
	for {
		if !duringPulse {
			took := time.Since(eventStart)
			if took > conf.timeout && pulsesThisTime != 0 {
				fmt.Printf("%s--- Pulses: %d%s\n", yellow, pulsesThisTime, reset)
				pulsesThisTime = 0
			}
		}

		// Pause is over.
		if p.Read() == gpio.Low {
			if !duringPulse {
				took := time.Since(eventStart)
				if took <= conf.timeout {
					fmt.Printf("Pause width: %d ms (%d μs)\n",
						took.Milliseconds(), took.Microseconds())
					totalPausesLength += took
					totalPausesNumber++
				}

				// Reset timer and start counting pulse width.
				eventStart = time.Now()
				duringPulse = true
			}
			continue
		}

		// Pulse is over.
		if duringPulse {
			took := time.Since(eventStart)
			totalPulsesLength += took
			totalPulsesNumber++
			pulsesThisTime++

			fmt.Printf("%sPulse width: %d ms (%d μs)%s\n", green,
				took.Milliseconds(), took.Microseconds(), reset)

			// Reset timer and start counting the pause width.
			eventStart = time.Now()
			duringPulse = false
		}
	}
}
