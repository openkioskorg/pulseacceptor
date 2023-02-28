package main

import (
	"log"

	"github.com/adjust/rmq/v5"
)

type redisConfig struct {
	Tag string `yaml:"tag"`

	// Connection type: tcp or unix
	Network string `yaml:"network"`
	Address  string `yaml:"address"`
	Db int `yaml:"db"`

	// Name of the queue
	Queue string `yaml:"queue"`
}

func initQueue(conf redisConfig) rmq.Queue {
	conn, err := rmq.OpenConnection(conf.Tag, conf.Network, conf.Address, conf.Db, nil)
	if err != nil {
		log.Fatal("Failed to connect to Redis: ", err)
	}
	queue, err := conn.OpenQueue(conf.Queue)
	if err != nil {
		log.Fatal("Failed to open acceptors queue: ", err)
	}
	return queue
}
