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

func (eh *EthernetHeader) Marshal() []byte {
	d := make([]byte, 14)
	copy(d[0:6], eh.addr.DstAddr)
	copy(d[6:12], eh.addr.SrcAddr)
	d[12] = byte(eh.EtherType >> 8)
	d[13] = byte(eh.EtherType)
	return d
}

func (eh *EthernetHeader) GetType() netstack.ProtocolType {
	return netstack.ProtocolTypeEthernet
}

func (eh *EthernetHeader) GetDstMAC() net.HardwareAddr {
	return eh.addr.DstAddr
}

func (eh *EthernetHeader) GetSrcMAC() net.HardwareAddr {
	return eh.addr.SrcAddr
}

/*
	Helper functions to convert between EtherType and netstack.ProtocolType
*/

func (eh *EthernetHeader) GetL3Type() netstack.ProtocolType {
	switch eh.EtherType {
	case EthernetTypeIPv4:
		return netstack.ProtocolTypeIPv4
	case EthernetTypeARP:
		return netstack.ProtocolTypeARP
	case EthernetTypeIPv6:
		return netstack.ProtocolTypeIPv6
	default:
		return netstack.ProtocolTypeUnknown
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