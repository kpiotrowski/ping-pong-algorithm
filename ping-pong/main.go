package main

import (
	"flag"
	"fmt"
	"math"
	"time"

	"github.com/kpiotrowski/ping-pong-algorithm/ping-pong/token"
	zmq "github.com/pebbe/zmq4"
	log "github.com/sirupsen/logrus"
)

const logLevel = log.InfoLevel

var (
	generateToken = flag.Bool("i", false, "Set to true if this node should send the first Ping")
	losePingRound = flag.Int("pi", 0, "Set to X>0 if you want to force a node to lose a Ping in the round X")
	losePongRound = flag.Int("po", 0, "Set to X>0 if you want to force a node to lose a Pong in the round X")
)

func main() {
	flag.Parse()
	log.SetLevel(logLevel)

	if len(flag.Args()) < 2 {
		panic("Not enough arguments to run. You should execute cohort with (flags) [node_addr:port] [next_node_addr:port]")
	}

	c, err := newNode(flag.Args()[0], flag.Args()[1])
	if err != nil {
		panic(err)
	}

	c.Run()
}

type node struct {
	lastToken      int
	nodeAddr       string
	nextNodeAddr   string
	nextNodeScoket *zmq.Socket
	msgChannel     chan token.Token
	pingToken      *token.Token
	pongToken      *token.Token
	round          int
	CSbusy         bool
	CSChannel      chan bool
	wantsCS        bool
}

func newNode(nodeAddr, nextAddr string) (*node, error) {
	n := &node{
		lastToken:    0,
		nodeAddr:     nodeAddr,
		nextNodeAddr: nextAddr,
		msgChannel:   make(chan token.Token),
		CSChannel:    make(chan bool),
		CSbusy:       false,
		wantsCS:      false,
		round:        0,
	}
	var err error
	n.nextNodeScoket, err = zmq.NewSocket(zmq.PUSH)
	if err != nil {
		return nil, err
	}
	err = n.nextNodeScoket.Connect(fmt.Sprintf("tcp://%s", nextAddr))

	return n, err
}

func (n *node) Run() {
	go token.Listen(n.msgChannel, n.nodeAddr)
	go n.triggerWantsCS()
	if *generateToken {
		n.sendFirstToken()
	}

	for {
		select {
		case _ = <-n.CSChannel:
			log.Warn(fmt.Sprintf("%s leaves Critical section", n.nodeAddr))
			n.CSbusy = false
			if n.pingToken != nil && n.pongToken != nil {
				n.incarnate(n.pingToken.Value)
			}
			n.forward()
			go n.triggerWantsCS()
		case t := <-n.msgChannel:
			if t.Value < 0 {
				n.pongToken = &t
			} else {
				n.pingToken = &t
			}
			if t.Value == n.lastToken {
				n.regenerate(t.Value)
			}
			n.lastToken = t.Value
			if !n.CSbusy {
				if n.wantsCS && t.Value > 0 {
					log.Warn(fmt.Sprintf("%s enters Critical section", n.nodeAddr))
					n.CSbusy = true
					n.wantsCS = false
					go func() {
						time.Sleep(time.Second * 5)
						n.CSChannel <- true
					}()
				} else {
					n.forward()
				}
			}
		}
	}
}

func (n *node) sendFirstToken() {
	n.pingToken = &token.Token{Value: 1}
	n.pongToken = &token.Token{Value: -1}
	n.forward()
}

func (n *node) regenerate(value int) {
	if value < 0 {
		log.Warn("PING token is lost, regenerating it")
	} else {
		log.Warn("PONG token is lost, regenerating it")
	}
	log.Warn("Regenerating token")
	n.pingToken = &token.Token{Value: int(math.Abs(float64(value)))}
	n.pongToken = &token.Token{Value: -int(math.Abs(float64(value)))}
}

func (n *node) incarnate(value int) {
	log.Warn("New incarnation of tokens")
	n.pingToken = &token.Token{Value: 1 + int(math.Abs(float64(value)))}
	n.pongToken = &token.Token{Value: -(1 + int(math.Abs(float64(value))))}
}

func (n *node) forward() {
	if n.pingToken != nil {
		n.forwardPing()
	}
	if n.pongToken != nil {
		n.forwardPong()
	}

}

func (n *node) forwardPing() {
	n.lastToken = n.pingToken.Value
	if n.round == *losePingRound && *losePingRound > 0 {
		log.Error("Ommiting PING send - PING token is lost")
	} else {
		token.SendToken(*n.pingToken, n.nodeAddr, n.nextNodeAddr, n.nextNodeScoket)
	}
	n.pingToken = nil
	n.round++
}

func (n *node) forwardPong() {
	n.lastToken = n.pongToken.Value
	if n.round == *losePongRound && *losePongRound > 0 {
		log.Error("Ommiting POST send - PONG token is lost")
	} else {
		token.SendToken(*n.pongToken, n.nodeAddr, n.nextNodeAddr, n.nextNodeScoket)
	}
	n.pongToken = nil
}

func (n *node) triggerWantsCS() {
	time.Sleep(time.Second * 30)
	n.wantsCS = true
	log.Warn("Wants to enter CS")
}
