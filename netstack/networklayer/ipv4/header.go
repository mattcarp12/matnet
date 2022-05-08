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

var (
	ErrInvalidIPv4Header = errors.New("invalid IPv4 header")
	ErrTTLZero           = errors.New("TTL is zero")
	ErrInvalidCheckSum   = errors.New("invalid checksum")
)

func (h *IPv4Header) Marshal() []byte {
	// make byte buffer for IPv4 header
	b := make([]byte, IPv4HeaderSize)

	// version and IHL
	b[0] = (4 << 4) | (5 & 0x0f)

	// type of service
	b[1] = 0

	// total length
	binary.BigEndian.PutUint16(b[2:4], uint16(h.TotalLength))

	// identification
	binary.BigEndian.PutUint16(b[4:6], h.Identification)

	// flags
	b[6] = (h.Flags << 5) & 0xe0

	// fragment offset
	binary.BigEndian.PutUint16(b[6:8], h.FragmentOffset)

	// TTL
	b[8] = h.TTL

	// protocol
	b[9] = h.Protocol

	// header checksum
	binary.BigEndian.PutUint16(b[10:12], h.HeaderChecksum)

	// source IP
	copy(b[12:16], h.SourceIP.To4())

	// destination IP
	copy(b[16:20], h.DestinationIP.To4())

	return b
}

func (h *IPv4Header) Unmarshal(b []byte) error {
	// check length
	if len(b) < IPv4HeaderSize {
		return ErrInvalidIPv4Header
	}

	// version
	h.Version = (b[0] & 0xf0) >> 4
	if h.Version != 4 {
		return ErrInvalidIPv4Header
	}

	// IHL
	h.IHL = b[0] & 0x0f

	// Enforce IHL value of 5 since not supporting IP options
	if h.IHL != 5 {
		return ErrInvalidIPv4Header
	}

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

	// Check that the TTL is not zero
	if h.TTL == 0 {
		return ErrTTLZero
	}

	// protocol
	h.Protocol = b[9]

	// header checksum
	h.HeaderChecksum = binary.BigEndian.Uint16(b[10:12])

	// source IP
	h.SourceIP = net.IP(b[12:16])

	// destination IP
	h.DestinationIP = net.IP(b[16:20])

	// Check the checksum of the header
	if netstack.Checksum(b[0:IPv4HeaderSize]) != 0 {
		return ErrInvalidCheckSum
	}

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
	if h.Protocol == ProtocolICMP {
		return netstack.ProtocolTypeICMPv4
	} else if h.Protocol == ProtocolTCP {
		return netstack.ProtocolTypeTCP
	} else if h.Protocol == ProtocolUDP {
		return netstack.ProtocolTypeUDP
	} else {
		return netstack.ProtocolTypeUnknown
	}
}

func GetIPProtocolType(proto netstack.ProtocolType) uint8 {
	switch proto {
	case netstack.ProtocolTypeICMPv4:
		return ProtocolICMP
	case netstack.ProtocolTypeTCP:
		return ProtocolTCP
	case netstack.ProtocolTypeUDP:
		return ProtocolUDP
	default:
		return 0
	}
}
