package main

import (
	"log"
	"os"

	pa "gitlab.com/openkiosk/pulseacceptor"
	"gopkg.in/yaml.v3"
)

type daemonConfig struct {
	Device *pa.PulseAcceptorConfig `yaml:"device"`
	Values map[uint64]uint64       `yaml:"values"`
	Mqtt   brokerConfig            `yaml:"mqtt"`
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
