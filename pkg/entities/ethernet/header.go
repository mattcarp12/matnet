package ethernet

import (
	"errors"
	"github.com/mattcarp12/go-net/pkg/entities/protocols"
)

type EthernetHeader struct {
	addr      EthernetAddressPair
	EtherType uint16
}

const (
	EthernetTypeIPv4 = 0x0800
	EthernetTypeARP  = 0x0806
	EthernetTypeIPv6 = 0x86DD
)

var ErrInvalidEthernetHeader = errors.New("invalid ethernet header")

func ParseEthernetHeader(b []byte) (*EthernetHeader, error) {
	eh := &EthernetHeader{}
	if len(b) < 14 {
		return nil, ErrInvalidEthernetHeader
	}

	// Set HW address in header
	eh.addr.DstAddr = b[0:6]
	eh.addr.SrcAddr = b[6:12]
	eh.EtherType = uint16(b[12])<<8 | uint16(b[13]) // Is in big endian

	return eh, nil
}

func (eh *EthernetHeader) Marshal() ([]byte, error) {
	d := make([]byte, 14)
	copy(d[0:6], eh.addr.DstAddr)
	copy(d[6:12], eh.addr.SrcAddr)
	d[12] = byte(eh.EtherType >> 8)
	d[13] = byte(eh.EtherType)
	return d, nil
}

func (eh *EthernetHeader) GetType() protocols.ProtocolType {
	return protocols.ProtocolTypeEthernet
}

func (eh *EthernetHeader) GetAddressPair() EthernetAddressPair {
	return eh.addr
}

func (eh *EthernetHeader) AppendToPDU(pdu protocols.PDU) error {
	headerBytes, err := eh.Marshal()
	if err != nil {
		return err
	}

	pdu.AppendBytes(headerBytes)

	return nil
}

func (eh *EthernetHeader) StripFromPDU(pdu protocols.PDU) {
	pdu.StripBytes(14)
}

func (eh *EthernetHeader) GetNextLayerType() (protocols.ProtocolType, error) {
	switch eh.EtherType {
	case EthernetTypeIPv4:
		return protocols.ProtocolTypeIPv4, nil
	case EthernetTypeARP:
		return protocols.ProtocolTypeARP, nil
	case EthernetTypeIPv6:
		return protocols.ProtocolTypeIPv6, nil
	default:
		return protocols.ProtocolUnknown, protocols.ErrProtocolNotFound
	}
}
