package core

import (
	"log"
	"net/url"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/taurusgroup/multi-party-sig/pkg/party"
	"github.com/taurusgroup/multi-party-sig/pkg/protocol"
)

func HandlerLoopByQueue(id party.ID, h protocol.Handler, network *WSNetwork) {
	// 从ws中获得数据，然后处理
	for {
		select {
		case msg, ok := <-h.Listen():
			if !ok {
				// close channel
				// network.done <- struct{}{}
				// log.Printf("close channel ID:%s", id.Domain())
				return
			}
			go network.Send(msg)
		case msg := <-network.inChannel:
			h.Accept(msg)
		}
	}
}

type WSNetwork struct {
	partyID    string
	done       chan struct{}
	mtx        sync.Mutex
	inChannel  chan *protocol.Message
	outChannel chan *protocol.Message
}

func NewWSeNetwork(partyID string) *WSNetwork {
	c := &WSNetwork{
		partyID:    partyID,
		done:       make(chan struct{}, 10),
		mtx:        sync.Mutex{},
		inChannel:  make(chan *protocol.Message, 1000),
		outChannel: make(chan *protocol.Message, 1000),
	}
	return c
}

func (n *WSNetwork) Init() {

	go func() {
		// 连接到websocket,监控数据
		var addr = "localhost:8080"
		u := url.URL{Scheme: "ws", Host: addr, Path: "/ws"}
		c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
		if err != nil {
			log.Fatal("dial:", err)
		}

		go func() {
			// 读取信息
			defer close(n.inChannel)
			defer close(n.outChannel)
			defer close(n.done)
			for {
				_, message, err := c.ReadMessage()
				if err != nil {
					log.Println("read:", err)
					return
				}
				protocolMsg := &protocol.Message{}
				protocolMsg.UnmarshalBinary(message)
				n.inChannel <- protocolMsg
			}

		}()

		for {
			// 写数据
			select {
			case <-n.done:
				return
			case msg := <-n.outChannel:
				msgBytes, _ := msg.MarshalBinary()
				err := c.WriteMessage(websocket.BinaryMessage, msgBytes)
				if err != nil {
					log.Println("write failed", err)
				}

			}
		}

	}()
}

func (n *WSNetwork) Send(msg *protocol.Message) {
	n.outChannel <- msg
}

func (n *WSNetwork) Quit(id party.ID) {
	n.mtx.Lock()
	defer n.mtx.Unlock()
}

func (n *WSNetwork) Done(id party.ID) chan struct{} {
	n.mtx.Lock()
	defer n.mtx.Unlock()
	close(n.done)
	return n.done
}
