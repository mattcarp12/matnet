package headers

import "github.com/mattcarp12/go-net/pkg/entities/protocols"

type Header interface {
	Marshal() ([]byte, error)
	GetType() protocols.ProtocolType
	AppendToPDU(pdu protocols.PDU) error
	StripFromPDU(pdu protocols.PDU)
	GetNextLayerType() (protocols.ProtocolType, error)
}

type HeaderParser func([]byte) (Header, error)
