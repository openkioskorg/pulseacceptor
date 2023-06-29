/* Library for counting pulses from money acceptors.
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

package pulseacceptor

import (
	"errors"
	"log"
	"time"

	"periph.io/x/conn/v3/gpio"
	"periph.io/x/conn/v3/gpio/gpioreg"
	"periph.io/x/conn/v3/gpio/gpioutil"
	"periph.io/x/conn/v3/physic"
)

var PinNotFound = errors.New("Pin not found")

type PulseAcceptorConfig struct {
	// The pulse pin.
	PulsePin string `yaml:"pulse_pin"`

	// The duration of pause width. You might want to add up an error margin on top of this value.
	Debounce time.Duration

	// The duration of pulse width. You might want to add up an error margin on top of this value.
	Denoise time.Duration

	// After how long of a pause, consider counting for this coin/bill done.
	Timeout time.Duration

	// Edge poll frequency
	Freq int `yaml:"poll_frequency"`

	// Print pulse and pause width/count logs
	Debug bool
}

type PulseAcceptorDevice struct {
	gpio.PinIO

	// Marks end of coin or bill insertion.
	Timeout time.Duration

	debug bool
}

// Returns a new device with a denoised and debounced input pin.
func Init(conf *PulseAcceptorConfig) (*PulseAcceptorDevice, error) {
	pin := gpioreg.ByName(conf.PulsePin)
	if pin == nil {
		return nil, PinNotFound
	}
	pin = gpioutil.PollEdge(pin, physic.Frequency(conf.Freq)*physic.Hertz)
	if err := pin.In(gpio.PullUp, gpio.BothEdges); err != nil {
		log.Fatal(err)
	}
	if conf.Debounce != 0 && conf.Denoise != 0 {
		var err error
		pin, err = gpioutil.Debounce(pin, conf.Denoise, conf.Debounce, gpio.BothEdges)
		if err != nil {
			return nil, err
		}
	}

	d := &PulseAcceptorDevice{PinIO: pin,
		debug:   conf.Debug,
		Timeout: conf.Timeout,
	}
	return d, nil
}

// Count pulses and return whenever the current wave of pulses ends.
// The debounce duration is used as the timeout value.
func (d *PulseAcceptorDevice) Count() (pulses int64) {
	eventStart := time.Now()
	duringPulse := false
	for {
		if !duringPulse {
			took := time.Since(eventStart)
			if took > d.Timeout && pulses != 0 {
				if d.debug {
					log.Printf("--- Pulses: %d\n", pulses)
				}
				return
			}
		}

		// Pause is over.
		if d.Read() == gpio.Low {
			if !duringPulse {
				took := time.Since(eventStart)
				if d.debug && took <= d.Timeout {
					log.Printf("Pause width: %d ms (%d μs)\n",
						took.Milliseconds(), took.Microseconds())
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
			pulses++

			if d.debug {
				log.Printf("Pulse width: %d ms (%d μs)\n",
					took.Milliseconds(), took.Microseconds())
			}

			// Reset timer and start counting the pause width.
			eventStart = time.Now()
			duringPulse = false
		}
	}
}

// Count pulses and shove them down a channel when activity occurs.
func (d *PulseAcceptorDevice) CountWithHandler(ch chan<- int64) {
	var pulses int64
	eventStart := time.Now()
	duringPulse := false
	for {
		if !duringPulse {
			took := time.Since(eventStart)
			if took > d.Timeout && pulses != 0 {
				ch <- pulses
				if d.debug {
					log.Printf("--- Pulses: %d\n", pulses)
				}
				pulses = 0
			}
		}

		// Pause is over.
		if d.Read() == gpio.Low {
			if !duringPulse {
				took := time.Since(eventStart)
				if d.debug && took <= d.Timeout {
					log.Printf("Pause width: %d ms (%d μs)\n",
						took.Milliseconds(), took.Microseconds())
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
			pulses++

			if d.debug {
				log.Printf("Pulse width: %d ms (%d μs)\n",
					took.Milliseconds(), took.Microseconds())
			}

			// Reset timer and start counting the pause width.
			eventStart = time.Now()
			duringPulse = false
		}
	}
}
