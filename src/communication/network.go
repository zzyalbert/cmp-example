package communication

import (
	"log"
	"net/url"
	"sync"

	"github.com/0x1be20/cmp-example/src/common"
	"github.com/gorilla/websocket"
	"github.com/taurusgroup/multi-party-sig/pkg/party"
	"github.com/taurusgroup/multi-party-sig/pkg/protocol"
)

type ProtocolNetwork interface {
	Send(msg *protocol.Message)
	Quit(id party.ID)
	Done(id party.ID) chan struct{}
	Next() chan *protocol.Message
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

func (n *WSNetwork) Init(addr string, path string) {

	// 连接到websocket,监控数据
	u := url.URL{Scheme: "ws", Host: addr, Path: path}
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
			protocolMsg := &common.Message{
				SessionId: "",
				ExtraData: "",
				Data:      &protocol.Message{},
			}
			protocolMsg.UnmarshalBinary(message)
			n.inChannel <- protocolMsg.Data
		}

	}()

	go func() {
		for {
			// 写数据
			select {
			case <-n.done:
				return
			case msg := <-n.outChannel:
				outMesg := &common.Message{
					SessionId: "",
					ExtraData: "",
					Data:      msg,
				}

				msgBytes, _ := outMesg.MarshalBinary()
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

func (n *WSNetwork) Next() chan *protocol.Message {
	return n.inChannel
}
