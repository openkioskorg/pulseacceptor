package main

import (
	"fmt"
	"log"
	"periph.io/x/conn/v3/gpio"
	"periph.io/x/conn/v3/gpio/gpioreg"
	"periph.io/x/conn/v3/gpio/gpioutil"
	"periph.io/x/conn/v3/physic"
	"periph.io/x/host/v3"
	"time"
)

var green = "\033[32m"
var reset = "\033[0m"

func main() {
	if _, err := host.Init(); err != nil {
		log.Fatal(err)
	}

	p := gpioreg.ByName("17")

	p = gpioutil.PollEdge(p, 10*physic.Hertz)
	if err := p.In(gpio.PullUp, gpio.BothEdges); err != nil {
		log.Fatal(err)
	}

	eventStart := time.Now()
	duringPulse := false
	for {
		// Pause is over.
		if p.Read() == gpio.High {
			if !duringPulse {
				took := time.Since(eventStart)
				fmt.Printf("Pause width: %d ms (%d μs)\n",
					took.Milliseconds(), took.Microseconds())

				// Reset timer and start counting pulse width.
				eventStart = time.Now()
				duringPulse = true
			}
			continue
		}

		// Pulse is over.
		if duringPulse {
			took := time.Since(eventStart)
			fmt.Printf("%sPulse width: %d ms (%d μs)%s\n", green,
				took.Milliseconds(), took.Microseconds(), reset)

			// Reset timer and start counting the pause width.
			eventStart = time.Now()
			duringPulse = false
		}
	}
}
