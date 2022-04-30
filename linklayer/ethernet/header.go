package ethernet

import (
	"errors"
	"net"

	"github.com/mattcarp12/go-net/netstack"
)

type EthernetAddressPair struct {
	DstAddr net.HardwareAddr
	SrcAddr net.HardwareAddr
}

type EthernetHeader struct {
	addr      EthernetAddressPair
	EtherType uint16
}

// Ethertype values
const (
	EthernetTypeIPv4 = 0x0800
	EthernetTypeARP  = 0x0806
	EthernetTypeIPv6 = 0x86DD
)

const EthernetHeaderSize = 14

var ErrInvalidEthernetHeader = errors.New("invalid ethernet header")

func (eh *EthernetHeader) Unmarshal(b []byte) error {
	if len(b) < EthernetHeaderSize {
		return ErrInvalidEthernetHeader
	}

	// Set HW address in header
	eh.addr.DstAddr = b[0:6]
	eh.addr.SrcAddr = b[6:12]
	eh.EtherType = uint16(b[12])<<8 | uint16(b[13]) // Is in big endian

	return nil
}

func (eh *EthernetHeader) Marshal() ([]byte, error) {
	d := make([]byte, 14)
	copy(d[0:6], eh.addr.DstAddr)
	copy(d[6:12], eh.addr.SrcAddr)
	d[12] = byte(eh.EtherType >> 8)
	d[13] = byte(eh.EtherType)
	return d, nil
}

func (eh *EthernetHeader) GetType() netstack.ProtocolType {
	return netstack.ProtocolTypeEthernet
}

func (eh *EthernetHeader) GetAddressPair() EthernetAddressPair {
	return eh.addr
}

func (eh *EthernetHeader) GetNextLayerType() (netstack.ProtocolType, error) {
	switch eh.EtherType {
	case EthernetTypeIPv4:
		return netstack.ProtocolTypeIPv4, nil
	case EthernetTypeARP:
		return netstack.ProtocolTypeARP, nil
	case EthernetTypeIPv6:
		return netstack.ProtocolTypeIPv6, nil
	default:
		return netstack.ProtocolTypeUnknown, netstack.ErrProtocolNotFound
	}
}

func GetEtherTypeFromProtocolType(pt netstack.ProtocolType) (uint16, error) {
	switch pt {
	case netstack.ProtocolTypeIPv4:
		return EthernetTypeIPv4, nil
	case netstack.ProtocolTypeARP:
		return EthernetTypeARP, nil
	case netstack.ProtocolTypeIPv6:
		return EthernetTypeIPv6, nil
	default:
		return 0, netstack.ErrProtocolNotFound
	}
}