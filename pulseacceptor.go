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
	"time"

	"periph.io/x/conn/v3/gpio"
	"periph.io/x/conn/v3/gpio/gpioreg"
	"periph.io/x/conn/v3/gpio/gpioutil"
)

var PinNotFound = errors.New("Pin not found")

type PulseAcceptorConfig struct {
	// The pulse pin.
	Pin string
	// The duration of pause width. You might want to add up an error margin on top of this value.
	Debounce time.Duration
	// The duration of pulse width. You might want to add up an error margin on top of this value.
	Denoise time.Duration
}

type PulseAcceptorDevice struct {
	gpio.PinIO

	// Marks end of coin or bill insertion.
	Timeout time.Duration
}

// Returns a new device with a denoised and debounced input pin.
func Init(conf *PulseAcceptorConfig) (*PulseAcceptorDevice, error) {
	pin := gpioreg.ByName(conf.Pin)
	if pin == nil {
		return nil, PinNotFound
	}
	d := &PulseAcceptorDevice{}
	var err error
	d.PinIO, err = gpioutil.Debounce(pin, conf.Denoise, conf.Debounce, gpio.BothEdges)
	if err != nil {
		return nil, err
	}
	d.Timeout = conf.Debounce
	return d, nil
}

// Count pulses and return whenever the current wave of pulses ends.
// The debounce duration is used as the timeout value.
func (d *PulseAcceptorDevice) Count() (pulses int64) {
	for {
		pulses = 0
		for d.WaitForEdge(d.Timeout) {
			if d.Read() == gpio.High {
				pulses++
			}
		}
		return
	}
}

// Count pulses and shove them down a channel when activity occurs.
func (d *PulseAcceptorDevice) CountWithHandler(ch chan<- int64) {
	var pulses int64
	for {
		pulses = 0
		for d.WaitForEdge(d.Timeout) {
			if d.Read() == gpio.High {
				pulses++
			}
		}
		if pulses != 0 {
			ch <- pulses
		}
	}
}
