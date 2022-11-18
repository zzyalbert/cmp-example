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

var serveNode = &websocket.Conn{}
var auditorNode = &websocket.Conn{}
var userNodes = make(map[string]*websocket.Conn)

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
			sessionId = protocolMesg.SessionId

			if protocolMesg.NodeType == common.NodeTypeServer {
				serveNode = c
			} else if protocolMesg.NodeType == common.NodeTypeAuditor {
				auditorNode = c
			} else {
				// user node
				userNodes[sessionId] = c
			}

			log.Printf("register %s node,session:%s id:%s", protocolMesg.NodeType, sessionId, id)

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
		sessionId = protocolMesg.SessionId
		log.Printf("session %s,recv msg, from %s,to %s round %v", sessionId, protocolMesg.Data.From, protocolMesg.Data.To, protocolMesg.Data.RoundNumber)

		err = serveNode.WriteMessage(mt, message)
		if err != nil {
			log.Printf("server node err,%+v", err)
		}

		err = auditorNode.WriteMessage(mt, message)
		if err != nil {
			log.Printf("auditor node err ,%+v", err)
		}

		err = userNodes[sessionId].WriteMessage(mt, message)
		if err != nil {
			log.Printf("user node err,%+v", err)
		}

		locker.Unlock()
	}
}

func InitServer(host string, port int64, path string) {
	http.HandleFunc(path, connect)
	var addr = fmt.Sprintf("%s:%d", host, port)
	log.Printf("address %s %s", addr, path)
	log.Fatal(http.ListenAndServe(addr, nil))
}
