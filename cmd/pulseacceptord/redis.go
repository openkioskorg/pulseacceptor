package main

import (
	"log"

	"github.com/adjust/rmq/v5"
)

type redisConfig struct {
	// Connection type: tcp or unix
	ConnType string `yaml:"type"`
	Host     string `yaml:"host"`
}

func initQueue(conf redisConfig) rmq.Queue {
	conn, err := rmq.OpenConnection("pulseacceptord", conf.ConnType, conf.Host, 1, nil)
	if err != nil {
		log.Fatal("Failed to connect to Redis: ", err)
	}
	queue, err := conn.OpenQueue("acceptors")
	if err != nil {
		log.Fatal("Failed to open acceptors queue: ", err)
	}
	return queue
}
