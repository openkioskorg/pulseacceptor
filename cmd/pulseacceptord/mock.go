package main

import (
	"math/rand"
	"time"
)

func bsvalues(ch chan<- uint64) {
	pulses := []uint64{10, 20, 50, 100, 200}
	for {
		ch <- pulses[rand.Intn(4)]
		time.Sleep(3 * time.Second)
	}
}
