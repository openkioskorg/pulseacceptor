package main

import (
	"encoding/json"
	"log"
	"context"
	"fmt"
	"time"
	"math/rand"
	"net/url"

	//pa "gitlab.com/openkiosk/pulseacceptor"
	"periph.io/x/host/v3"
	"github.com/eclipse/paho.golang/autopaho"
	"github.com/eclipse/paho.golang/paho"
)

type pulseAcceptorEvent struct {
	Amount uint64 `json:"amount_from_2"`
}

func bsvalues(ch chan<- uint64) {
	pulses := []uint64{10,20,50,100,200}
	for {
		ch <- pulses[rand.Intn(5)]
		time.Sleep(3 * time.Second)
	}
}

func subHandler(msg *paho.Publish) {
	log.Println("!!!Received!!! ", string(msg.Payload))
}

func main() {
	conf := parseConfig()
	if _, err := host.Init(); err != nil {
		log.Fatal("Failed to load host drivers: ", err)
	}

/*	pulseDevice, err := pa.Init(conf.Device)
	if err != nil {
		log.Fatal("Failed to initialize pulse device: ", err)
	}
*/
	u, _ := url.Parse("mqtt://127.0.0.1:1883")
	broker, err := autopaho.NewConnection(context.Background(), autopaho.ClientConfig{
		BrokerUrls: []*url.URL{u},
		OnConnectionUp: func(cm *autopaho.ConnectionManager, connAck *paho.Connack) {
			fmt.Println("mqtt connection up")
			if _, err := cm.Subscribe(context.Background(), &paho.Subscribe{
				Subscriptions: map[string]paho.SubscribeOptions{
					"pulseacceptord": {QoS: 2, NoLocal: true},
				},
			}); err != nil {
				fmt.Printf("failed to subscribe (%s). This is likely to mean no messages will be received.", err)
				return
			}
			fmt.Println("mqtt subscription made")
		},
		OnConnectError:    func(err error) { fmt.Printf("error whilst attempting connection: %s\n", err) },
		ClientConfig: paho.ClientConfig{
			ClientID: "pulseacceptord-1",
			Router: paho.NewSingleHandlerRouter(func(m *paho.Publish) {
				subHandler(m)
			}),
			OnClientError: func(err error) { fmt.Printf("server requested disconnect: %s\n", err) },
			OnServerDisconnect: func(d *paho.Disconnect) {
				if d.Properties != nil {
					fmt.Printf("server requested disconnect: %s\n", d.Properties.ReasonString)
				} else {
					fmt.Printf("server requested disconnect; reason code: %d\n", d.ReasonCode)
				}
			},
		},
	})

	if err != nil {
		log.Fatal("Failed to connect to MQTT broker: ", err)
	}

	pulseChan := make(chan uint64)
	//go pulseDevice.CountWithHandler(pulseChan)
	go bsvalues(pulseChan)

	for {
		select {
		case p := <-pulseChan:
			amount := conf.Values[p]
			log.Printf("Received %d cents.\n", amount)
			publishAmount(context.Background(), broker, amount)/*; err != nil {
				log.Println("Failed to publish event: ", err)
			}*/
		}
	}
}

func publishAmount(ctx context.Context, broker *autopaho.ConnectionManager, amount uint64) {
	// AwaitConnection will return immediately if connection is up; adding this call stops publication whilst
	// connection is unavailable.
	if err := broker.AwaitConnection(ctx); err != nil { // Should only happen when context is cancelled
		fmt.Printf("publisher done (AwaitConnection: %s)\n", err)
		return
	}

	msg, err := json.Marshal(pulseAcceptorEvent{Amount: amount})
	if err != nil {
		log.Fatal(err)
	}

	// Publish will block so we run it in a goroutine
	go func(msg []byte) {
		pr, err := broker.Publish(ctx, &paho.Publish{
			QoS:     2,
			Topic:   "pulseacceptord-2",
			Payload: msg,
		})
		if err != nil {
			fmt.Printf("error publishing: %s\n", err)
		} else if pr.ReasonCode != 0 && pr.ReasonCode != 16 { // 16 = Server received message but there are no subscribers
			fmt.Printf("reason code %d received\n", pr.ReasonCode)
		}
		fmt.Printf("sent message: %s\n", msg)
	}(msg)
}
