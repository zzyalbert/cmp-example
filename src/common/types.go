package common

import (
	"github.com/fxamacker/cbor/v2"
	"github.com/taurusgroup/multi-party-sig/pkg/protocol"
)

type Message struct {
	Type      MesgType
	SessionId string
	MesgId    int64
	RequestId string
	ExtraData []byte
	Data      *protocol.Message
}

type marshallableMessage struct {
	Type      string
	SessionId string
	MesgId    int64
	RequestId string
	ExtraData []byte
	Data      []byte
}

func (m *Message) toMarshallable() *marshallableMessage {
	protocolData, _ := m.Data.MarshalBinary()
	return &marshallableMessage{
		Type:      string(m.Type),
		SessionId: m.SessionId,
		MesgId:    m.MesgId,
		RequestId: m.RequestId,
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
	cbor.Unmarshal(deserialized.Data, &m.Data)
	m.ExtraData = deserialized.ExtraData
	m.SessionId = deserialized.SessionId
	m.Type = MesgType(deserialized.Type)
	m.RequestId = deserialized.RequestId
	m.MesgId = deserialized.MesgId
	return nil
}

type MesgType string

const (
	MesgTypeRegister   MesgType = "register"
	MesgTypeProtocol   MesgType = "protocol"
	MesgTypeCommon     MesgType = "common"
	MesgTypeSign       MesgType = "sign"
	MesgTypeSignResult MesgType = "result_sign"
	MesgTypeReqSign    MesgType = "req_sign"
)

func (this MesgType) String() string {
	return string(this)
}

type Signature struct {
	X []byte
	S []byte
}

type marshallableSignature struct {
	X []byte
	S []byte
}

func (s *Signature) toMarshallable() *marshallableSignature {
	return &marshallableSignature{
		X: s.X,
		S: s.S,
	}
}

func (s *Signature) MarshalBinary() ([]byte, error) {
	return cbor.Marshal(s.toMarshallable())
}

func (s *Signature) Unmarshal(data []byte) error {
	if err := cbor.Unmarshal(data, s); err != nil {
		return err
	}
	return nil
}
