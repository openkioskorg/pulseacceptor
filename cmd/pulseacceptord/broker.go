package main

import (
	"context"
	"encoding/json"
	"log"
	"net/url"

	"github.com/eclipse/paho.golang/autopaho"
	"github.com/eclipse/paho.golang/paho"
)

type brokerConfig struct {
	Brokers  []string `yaml:"brokers"`
	Topic    string   `yaml:"topic"`
	ClientID string   `yaml:"client_id"`
}

type mqttBroker struct {
	topic string
	*autopaho.ConnectionManager
}

func newBroker(conf brokerConfig) (*mqttBroker, error) {
	var brokerUrls []*url.URL
	for _, urlStr := range conf.Brokers {
		u, err := url.Parse(urlStr)
		if err != nil {
			return nil, err
		}
		brokerUrls = append(brokerUrls, u)
	}

	b := &mqttBroker{}
	cm, err := autopaho.NewConnection(context.Background(), autopaho.ClientConfig{
		BrokerUrls: brokerUrls,
		OnConnectionUp: func(cm *autopaho.ConnectionManager, connAck *paho.Connack) {
			log.Println("MQTT connection up")
			if _, err := cm.Subscribe(context.Background(), &paho.Subscribe{
				Subscriptions: map[string]paho.SubscribeOptions{
					conf.Topic: {QoS: 2, NoLocal: true},
				},
			}); err != nil {
				log.Printf("Failed to subscribe (%s).", err)
				return
			}
			log.Println("MQTT subscription made")
		},
		OnConnectError: func(err error) { log.Printf("Error whilst attempting connection: %s\n", err) },
		ClientConfig: paho.ClientConfig{
			ClientID: conf.ClientID,
			Router: paho.NewSingleHandlerRouter(func(m *paho.Publish) {
				commandHandler(m)
			}),
			OnClientError: func(err error) { log.Printf("Server requested disconnect: %s\n", err) },
			OnServerDisconnect: func(d *paho.Disconnect) {
				if d.Properties != nil {
					log.Printf("Server requested disconnect: %s\n", d.Properties.ReasonString)
				} else {
					log.Printf("Server requested disconnect; reason code: %d\n", d.ReasonCode)
				}
			},
		},
	})
	b.ConnectionManager = cm
	b.topic = conf.Topic
	return b, err
}

func (b *mqttBroker) publishAmount(ctx context.Context, amount uint64) error {
	// AwaitConnection will return immediately if connection is up; adding this call stops publication whilst
	// connection is unavailable.
	if err := b.AwaitConnection(ctx); err != nil { // Should only happen when context is cancelled
		return err
	}

	msg, err := json.Marshal(PulseacceptordEvent{Amount: amount})
	if err != nil {
		return err
	}

	// Publish will block so we run it in a goroutine
	go func(msg []byte) {
		pr, err := b.Publish(ctx, &paho.Publish{
			QoS:     2,
			Topic:   b.topic,
			Payload: msg,
		})
		if err != nil {
			log.Printf("Error publishing: %s\n", err)
		} else if pr.ReasonCode != 0 && pr.ReasonCode != 16 { // 16 = Server received message but there are no subscribers
			log.Printf("Reason code %d received\n", pr.ReasonCode)
		}
		log.Printf("Sent message: %s\n", msg)
	}(msg)
	return nil
}