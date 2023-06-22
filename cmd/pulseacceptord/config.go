/* Daemon for counting pulses from money acceptors.
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
	"log"
	"os"

	pa "gitlab.com/openkiosk/pulseacceptor"
	"gopkg.in/yaml.v3"
)

var conf daemonConfig

type daemonConfig struct {
	Device           *pa.PulseAcceptorConfig `yaml:"device"`
	Values           map[int64]int64         `yaml:"values"`
	Mqtt             brokerConfig            `yaml:"mqtt"`
	EnablePinControl bool                    `yaml:"enable_pin_control"`
	EnablePin        string                  `yaml:"enable_pin"`
	EnabledWhenHigh  bool                    `yaml:"enabled_when_high"`
}

func parseConfig(filename string) (conf daemonConfig) {
	file, err := os.ReadFile(filename)
	if err != nil {
		log.Fatal("Failed to read config: ", err)
	}
	if err := yaml.Unmarshal(file, &conf); err != nil {
		log.Fatal("Failed to unmarshal yaml: ", err)
	}
	return
}
