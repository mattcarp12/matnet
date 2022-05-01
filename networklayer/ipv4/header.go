package ipv4

import (
	"encoding/binary"
	"errors"
	"net"

	"github.com/mattcarp12/go-net/netstack"
)

type IPv4Header struct {
	Version        uint8
	IHL            uint8
	TypeOfService  uint8
	TotalLength    uint16
	Identification uint16
	Flags          uint8
	FragmentOffset uint16
	TTL            uint8
	Protocol       uint8
	HeaderChecksum uint16
	SourceIP       net.IP
	DestinationIP  net.IP
}

const (
	ProtocolICMP = 1
	ProtocolTCP  = 6
	ProtocolUDP  = 17
)

const IPv4HeaderSize = 20

var ErrInvalidIPv4Header = errors.New("invalid IPv4 header")

func (h *IPv4Header) Marshal() ([]byte, error) {
	// make byte buffer for IPv4 header
	b := make([]byte, IPv4HeaderSize)

	// version
	b[0] = (4 << 4) | (5 & 0x0f)

	// IHL
	b[1] = 5

	// type of service
	b[2] = 0

	// total length
	binary.BigEndian.PutUint16(b[3:5], uint16(h.TotalLength))

	// identification
	binary.BigEndian.PutUint16(b[5:7], h.Identification)

	// flags
	b[7] = 0

	// fragment offset
	binary.BigEndian.PutUint16(b[8:10], h.FragmentOffset)

	// TTL
	b[10] = h.TTL

	// protocol
	b[11] = h.Protocol

	// header checksum
	binary.BigEndian.PutUint16(b[12:14], h.HeaderChecksum)

	// source IP
	copy(b[14:18], h.SourceIP.To4())

	// destination IP
	copy(b[18:22], h.DestinationIP.To4())

	return b, nil
}

func (h *IPv4Header) Unmarshal(b []byte) error {
	// check length
	if len(b) < IPv4HeaderSize {
		return ErrInvalidIPv4Header
	}

	// version
	h.Version = (b[0] & 0xf0) >> 4

	// IHL
	h.IHL = b[0] & 0x0f

	// type of service
	h.TypeOfService = b[1]

	// total length
	h.TotalLength = binary.BigEndian.Uint16(b[2:4])

	// identification
	h.Identification = binary.BigEndian.Uint16(b[4:6])

	// flags
	h.Flags = b[6] >> 5

	// fragment offset
	h.FragmentOffset = binary.BigEndian.Uint16(b[6:8]) & 0x1fff

	// TTL
	h.TTL = b[8]

	// protocol
	h.Protocol = b[9]

	// header checksum
	h.HeaderChecksum = binary.BigEndian.Uint16(b[10:12])

	// source IP
	h.SourceIP = net.IP(b[12:16])

	// destination IP
	h.DestinationIP = net.IP(b[16:20])

	return nil
}

func (h *IPv4Header) GetType() netstack.ProtocolType {
	return netstack.ProtocolTypeIPv4
}

func (h *IPv4Header) GetSrcIP() net.IP {
	return h.SourceIP
}

func (h *IPv4Header) GetDstIP() net.IP {
	return h.DestinationIP
}

func (h *IPv4Header) GetL4Type() netstack.ProtocolType {
	// TODO: implement this
	return netstack.ProtocolType(h.Protocol)
}
