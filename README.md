# pulseacceptor
This component is for coin acceptors and bill acceptors that use the pulse interface.

`pulseacceptor.go` contains a small library for counting pulses from a GPIO pin. The input is denoised and debounced.\
`cmd/pulse-watcher` contains a program for checking the pulse and pause widths sent by the device.\
`cmd/pulseacceptord` is a daemon that counts pulses, translates them into monetary values and sends this information to Redis.
