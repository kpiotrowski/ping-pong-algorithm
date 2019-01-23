package token

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	zmq "github.com/pebbe/zmq4"
	log "github.com/sirupsen/logrus"
)

type Token struct {
	Value  int    `json:"value"`
	Sender string `json:"sender"`
}

func SendToken(token Token, sender, receiver string, socket *zmq.Socket) error {
	time.Sleep(time.Second)
	if token.Value < 0 {
		log.Info(fmt.Sprintf("%s sends PONG message with value %d to %s", sender, token.Value, receiver))
	} else {
		log.Info(fmt.Sprintf("%s sends PING message with value %d to %s", sender, token.Value, receiver))
	}

	token.Sender = sender
	payload, _ := json.Marshal(token)
	_, err := socket.SendBytes(payload, zmq.DONTWAIT)
	if err != nil {
		log.Error("Failed to send token", err)
	}
	return err
}

func Listen(deliverChannel chan Token, address string) {
	receiver, err := zmq.NewSocket(zmq.PULL)
	defer receiver.Close()
	if err != nil {
		log.Error("Failed to create socket: ", err)
	}

	receiver.Bind(fmt.Sprintf("tcp://*:%s", strings.Split(address, ":")[1]))
	for {
		data, err := receiver.RecvBytes(0)
		if err != nil {
			log.Error("Failed receive data: ", err)
			continue
		}

		message := Token{}
		err = json.Unmarshal(data, &message)
		if err != nil {
			log.Error("Failed to decode message", err)
		} else {
			if message.Value < 0 {
				log.Info(fmt.Sprintf("Received PONG message with value %d from %s", message.Value, message.Sender))
			} else {
				log.Info(fmt.Sprintf("Received PING message with value %d from %s", message.Value, message.Sender))
			}
			deliverChannel <- message
		}
	}
}
