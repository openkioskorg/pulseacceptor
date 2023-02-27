package main

import (
	"gopkg.in/yaml.v3"
	"os"
	"log"
	"periph.io/x/host/v3"
	pa "gitlab.com/openkiosk/pulseacceptor"
)

type daemonConfig struct {
	Device *pa.PulseAcceptorConfig `yaml:"device"`
	Values map[uint64]uint64 `yaml:"values"`
	Redis string `yaml:"redis"`
}

func parseConfig() (conf daemonConfig) {
	file, err := os.ReadFile("config.yaml")
	if err != nil {
		log.Fatal("Failed to read config: ", err)
	}
	if err := yaml.Unmarshal(file, &conf); err != nil {
		log.Fatal("Failed to unmarshal yaml: ", err)
	}
	return
}

func main() {
	conf := parseConfig()
	if _, err := host.Init(); err != nil {
                log.Fatal("Failed to load host drivers: ", err)
        }

	pulseDevice, err := pa.Init(conf.Device)
	if err != nil {
		log.Fatal("Failed to initialize pulse device: ", err)
	}

	pulseChan := make(chan uint64)
	go pulseDevice.CountWithHandler(pulseChan)

	for {
	select {
		case p := <-pulseChan:
			log.Printf("Received %d pulses.\n", p)
	}
}

/*	for {
	pulses := pulseDevice.Count()
	log.Printf("Received %d pulses.\n", pulses)
}*/

}
