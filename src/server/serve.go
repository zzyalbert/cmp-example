package server

import (
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/0x1be20/cmp-example/src/common"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/taurusgroup/multi-party-sig/pkg/protocol"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
} // use default options
var channels = map[string]map[string]*websocket.Conn{}
var clients = map[string]*websocket.Conn{} // requestid=>connect
var chLocker = sync.Mutex{}

var locker = sync.Mutex{}

func connect(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()
	id := uuid.New().String()
	log.Printf("connect %s", id)
	sessionId := ""

	for {
		mt, message, err := c.ReadMessage()
		if err != nil {
			log.Println("close channel:", id)
			// 断开链接
			chLocker.Lock()
			if sessionId != "" {
				delete(channels[sessionId], id)
			}
			chLocker.Unlock()
			break
		}
		protocolMesg := &common.Message{
			SessionId: id,
			ExtraData: []byte{},
			Data:      &protocol.Message{},
		}
		protocolMesg.UnmarshalBinary(message)
		locker.Lock()

		// router
		if protocolMesg.Type == common.MesgTypeRegister {
			chLocker.Lock()
			if _, ok := channels[sessionId]; ok {
				delete(channels[sessionId], id)
			}
			chLocker.Unlock()

			sessionId = protocolMesg.SessionId

			log.Printf("register node,session:%s id:%s", sessionId, id)
			if _, ok := channels[sessionId]; !ok {
				channels[sessionId] = make(map[string]*websocket.Conn)
				channels[sessionId][id] = c
			} else {
				channels[sessionId][id] = c
			}
			locker.Unlock()
			continue
		} else if protocolMesg.Type == common.MesgTypeReqSign {
			clients[protocolMesg.RequestId] = c
			// 转发到group中去
			log.Printf("request sign,ssessionid %s,reqid %s", protocolMesg.SessionId, protocolMesg.RequestId)

		} else if protocolMesg.Type == common.MesgTypeSignResult {
			// 转发给api
			log.Printf("got result")
			api := clients[protocolMesg.RequestId]
			err := api.WriteMessage(mt, message)
			if err != nil {
				log.Printf("write to api fail:%+v", err)
			}
			locker.Unlock()

			continue
		}

		log.Printf("session %s,recv msg, from %s,to %s round %v", sessionId, protocolMesg.Data.From, protocolMesg.Data.To, protocolMesg.Data.RoundNumber)
		sessionId = protocolMesg.SessionId

		chLocker.Lock()

		for _, conn := range channels[sessionId] {
			err = conn.WriteMessage(mt, message)
			if err != nil {
				log.Println("write:", err)
				break
			}
		}
		chLocker.Unlock()
		locker.Unlock()
	}
}

func InitServer(host string, port int64, path string) {
	http.HandleFunc(path, connect)
	var addr = fmt.Sprintf("%s:%d", host, port)
	log.Printf("address %s %s", addr, path)
	log.Fatal(http.ListenAndServe(addr, nil))
}
