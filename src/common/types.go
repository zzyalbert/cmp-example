package common

import (
	"github.com/fxamacker/cbor/v2"
	"github.com/taurusgroup/multi-party-sig/pkg/protocol"
)

type Message struct {
	SessionId string            `json:"session_id"`
	ExtraData string            `json:"extra_data"`
	Data      *protocol.Message `json:"protocol_data"`
}

type marshallableMessage struct {
	SessionId string
	ExtraData string
	Data      []byte
}

func (m *Message) toMarshallable() *marshallableMessage {
	protocolData, _ := m.Data.MarshalBinary()
	return &marshallableMessage{
		SessionId: m.SessionId,
		ExtraData: m.ExtraData,
		Data:      protocolData,
	}
}

func (m *Message) MarshalBinary() ([]byte, error) {
	return cbor.Marshal(m.toMarshallable())
}

func (m *Message) UnmarshalBinary(data []byte) error {
	deserialized := m.toMarshallable()
	if err := cbor.Unmarshal(data, deserialized); err != nil {
		return err
	}
	protocolMesg := protocol.Message{}
	cbor.Unmarshal(deserialized.Data, &protocolMesg)
	m.ExtraData = deserialized.ExtraData
	m.SessionId = deserialized.SessionId
	m.Data = &protocolMesg
	return nil
}
