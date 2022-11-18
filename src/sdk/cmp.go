package sdk

import (
	"log"

	"github.com/0x1be20/cmp-example/src/common"
	"github.com/0x1be20/cmp-example/src/communication"
	"github.com/taurusgroup/multi-party-sig/pkg/ecdsa"
	"github.com/taurusgroup/multi-party-sig/pkg/math/curve"
	"github.com/taurusgroup/multi-party-sig/pkg/party"
	"github.com/taurusgroup/multi-party-sig/pkg/pool"
	"github.com/taurusgroup/multi-party-sig/pkg/protocol"
	"github.com/taurusgroup/multi-party-sig/protocols/cmp"
)

// 该package简化使用CMP协议
//

func handleProtocol(h *protocol.MultiHandler, n communication.ProtocolNetwork, done chan int, stop chan int, requestId string, sessionId string) {
	go func() {
		defer close(done)
		defer close(stop)
		for {
			select {
			case msg, ok := <-h.Listen():
				if !ok {
					done <- 1
					return
				}
				cmsg := &common.Message{
					Type:      "protocol",
					SessionId: sessionId,
					RequestId: requestId,
					ExtraData: []byte{},
					Data:      msg,
				}
				// log.Printf("outcome mesg sessionid %s", cmsg.SessionId)
				go n.Send(cmsg)
			case msg := <-n.Next():
				if msg.Type == "protocol" {
					// log.Printf("income mesg sessionId %s", msg.SessionId)
					h.Accept(msg.Data)
				}
			case _ = <-stop:
				return
			}
		}
	}()

}

func CMPKeygen(id party.ID, ids party.IDSlice, threshold int, n communication.ProtocolNetwork, pl *pool.Pool, requestId string, sessionId string) (chan interface{}, chan int, error) {
	h, err := protocol.NewMultiHandler(cmp.Keygen(curve.Secp256k1{}, id, ids, threshold, pl), nil)
	done := make(chan int, 100)
	stop := make(chan int, 100)
	result := make(chan interface{}, 100)
	if err != nil {
		return result, stop, err
	}

	go handleProtocol(h, n, done, stop, requestId, sessionId)
	go func() {
		defer close(result)
		<-done
		r, err := h.Result()
		if err != nil {
			log.Printf("cmp keygen failed:%+v", err)
		}
		result <- r
	}()
	return result, stop, nil
}

func CMPSign(c *cmp.Config, m []byte, signers party.IDSlice, n communication.ProtocolNetwork, pl *pool.Pool, requestId string, sessionId string) (chan interface{}, chan int, error) {
	h, err := protocol.NewMultiHandler(cmp.Sign(c, signers, m, pl), nil)
	result := make(chan interface{}, 100)
	done := make(chan int, 100)
	stop := make(chan int, 100)
	if err != nil {
		return result, stop, err
	}

	go handleProtocol(h, n, done, stop, requestId, sessionId)

	go func() {
		defer close(result)
		<-done
		r, err := h.Result()
		if err != nil {
			log.Printf("cmp sign failed:%+v", err)
			return
		}

		signature := r.(*ecdsa.Signature)
		if !signature.Verify(c.PublicPoint(), m) {
			done <- -1
			log.Printf("failed to verify cmp signature")
			return
		}
		result <- signature
	}()

	return result, stop, nil
}
