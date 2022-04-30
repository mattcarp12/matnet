package arp

import (
	"encoding/binary"
	"errors"
	"net"

	"github.com/mattcarp12/go-net/netstack"
)

var (
	ErrInvalidHWAddr    = errors.New("invalid hardware address")
	ErrInvalidIPAddr    = errors.New("invalid ip address")
	ErrInvalidARPHeader = errors.New("invalid arp header")
)

const (
	ARP_REQUEST = 1
	ARP_REPLY   = 2
)

const ARP_HardwareTypeEthernet = 1
const ARP_ProtocolTypeIPv4 = 0x0800

type ARPHeader struct {
	HardwareType uint16
	ProtocolType uint16
	HardwareSize uint8
	ProtocolSize uint8
	OpCode       uint16
	SourceHWAddr net.HardwareAddr
	SourceIPAddr net.IP
	TargetHWAddr net.HardwareAddr
	TargetIPAddr net.IP
}

func (ah *ARPHeader) Unmarshal(b []byte) error {
	if len(b) < 8 {
		return ErrInvalidARPHeader
	}

	ah.HardwareType = binary.BigEndian.Uint16(b[0:2])
	ah.ProtocolType = binary.BigEndian.Uint16(b[2:4])
	ah.HardwareSize = b[4]
	ah.ProtocolSize = b[5]
	ah.OpCode = binary.BigEndian.Uint16(b[6:8])

	// Parse variable length addresses
	minLen := 8 + int(ah.HardwareSize)*2 + int(ah.ProtocolSize)*2
	if len(b) < minLen {
		return ErrInvalidARPHeader
	}

	// Source HW address
	ah.SourceHWAddr = b[8 : 8+int(ah.HardwareSize)]

	// Source IP address
	ah.SourceIPAddr = b[8+int(ah.HardwareSize) : 8+int(ah.HardwareSize)+int(ah.ProtocolSize)]

	// Target HW address
	ah.TargetHWAddr = b[8+int(ah.HardwareSize)+int(ah.ProtocolSize) : 8+int(ah.HardwareSize)*2+int(ah.ProtocolSize)]

	// Target IP address
	ah.TargetIPAddr = b[8+int(ah.HardwareSize)*2+int(ah.ProtocolSize) : 8+int(ah.HardwareSize)*2+int(ah.ProtocolSize)*2]

	return nil

}

func (arpHeader *ARPHeader) Marshal() ([]byte, error) {
	b := make([]byte, 8+len(arpHeader.SourceHWAddr)+len(arpHeader.SourceIPAddr)+len(arpHeader.TargetHWAddr)+len(arpHeader.TargetIPAddr))

	// Hardware type
	binary.BigEndian.PutUint16(b[0:2], arpHeader.HardwareType)

	// Protocol type
	binary.BigEndian.PutUint16(b[2:4], arpHeader.ProtocolType)

	// Hardware size
	b[4] = arpHeader.HardwareSize

	// Protocol size
	b[5] = arpHeader.ProtocolSize

	// Op code
	binary.BigEndian.PutUint16(b[6:8], uint16(arpHeader.OpCode))

	// Source HW address
	copy(b[8:8+int(arpHeader.HardwareSize)], arpHeader.SourceHWAddr)

	// Source IP address
	copy(b[8+int(arpHeader.HardwareSize):8+int(arpHeader.HardwareSize)+int(arpHeader.ProtocolSize)], arpHeader.SourceIPAddr)

	// Target HW address
	copy(b[8+int(arpHeader.HardwareSize)+int(arpHeader.ProtocolSize):8+int(arpHeader.HardwareSize)*2+int(arpHeader.ProtocolSize)], arpHeader.TargetHWAddr)

	// Target IP address
	copy(b[8+int(arpHeader.HardwareSize)*2+int(arpHeader.ProtocolSize):8+int(arpHeader.HardwareSize)*2+int(arpHeader.ProtocolSize)*2], arpHeader.TargetIPAddr)

	return b, nil
}

func (arpHeader *ARPHeader) GetDstIP() net.IP {
	return arpHeader.TargetIPAddr
}

func (arpHeader *ARPHeader) GetSrcIP() net.IP {
	return arpHeader.SourceIPAddr
}

func (arpHeader *ARPHeader) GetType() netstack.ProtocolType {
	return netstack.ProtocolTypeARP
}

func (arpHeader *ARPHeader) GetL4Type() netstack.ProtocolType {
	// This shouldn't be needed...
	return netstack.ProtocolTypeUnknown
}
