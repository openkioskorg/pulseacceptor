package pulseacceptor

import (
	"time"

	"periph.io/x/conn/v3/gpio"
	"periph.io/x/conn/v3/gpio/gpioreg"
	"periph.io/x/conn/v3/gpio/gpioutil"
)

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
func Init(conf *PulseAcceptorConfig) (d *PulseAcceptorDevice, err error) {
	pin := gpioreg.ByName(conf.Pin)
	d.PinIO, err = gpioutil.Debounce(pin, conf.Denoise, conf.Debounce, gpio.BothEdges)
	d.Timeout = conf.Debounce
	return
}

// Count pulses and return whenever the current wave of pulses ends.
// The debounce duration is used as the timeout value.
func (d *PulseAcceptorDevice) Count() (pulses uint64) {
	for {
		pulses = 0
		for d.WaitForEdge(d.Timeout) {
			if d.Read() == gpio.High {
				pulses++
			}
		}
		return pulses
	}
}

// Count pulses and shove them down a channel when activity occurs.
func (d *PulseAcceptorDevice) CountWithHandler(ch chan<- uint64) {
	var pulses uint64
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
