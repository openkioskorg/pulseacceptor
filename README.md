# pulseacceptor
This component is for coin acceptors and bill acceptors that use the pulse interface.

`pulseacceptor.go` contains a small library for counting pulses from a GPIO pin. The input is denoised and debounced.\
`cmd/pulse-watcher` contains a program for checking the pulse and pause widths sent by the device.\
`cmd/pulseacceptord` is a daemon that counts pulses, translates them into monetary values and sends this information to the MQTT broker.

## Debugging and finding config parameters using `pulse-watcher`
`pulse-watcher` is an useful program that prints widths of each pulse and pause. Upon exiting it also prints averages for both. You can use it to figure out values for debounce and denoise parameters for `pulseacceptord`.

Example usage:
```sh
# Listen on GPIO pin 17 for pulses, if no pulses follow after 101 milliseconds
# ignore this pause because pulses for this bill/coin has ended.
pulse-watcher -pin 17 -timeout 101ms
```

## Handling coins/cash with `pulseacceptord`
`pulseacceptord` denoises and debounces the pulses to prevent noise. After it is done counting pulses for a single bill/coin, it sends the event in JSON format to MQTT for queueing.

Change the values inside `config.yaml` before running.

**denoise**: Should be set to the pulse width. Although you might want to add or deduct a very little amount from this value to allow for some error. Pulse widths smaller than this value will be ignored.

**debounce**: Should be set to the pause width. Again, tune this value for possible errors like you should do for the `denoise` parameter. After a valid pulse (with width longer than the `denoise` parameter) has been received, more pulses will be ignored for the `debounce` amount. Meaning repeated pulses within `debounce` duration are ignored.
