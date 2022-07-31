package networklayer

import (
	"encoding/binary"
	"errors"
	"log"
	"net"
	"os"

	"github.com/mattcarp12/matnet/netstack"
)

var ip4_log = log.New(os.Stdout, "[IPv4] ", log.Ldate|log.Lmicroseconds|log.Lshortfile)

// =============================================================================
// IPv4 Header
// =============================================================================

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

func GetIPProtocolType(skb *netstack.SkBuff) (uint8, error) {
	l4Header, err := skb.GetL4Header()
	if err != nil {
		return 0, err
	}

	switch l4Header.GetType() {
	case netstack.ProtocolTypeICMPv4:
		return ProtocolICMP, nil
	case netstack.ProtocolTypeTCP:
		return ProtocolTCP, nil
	case netstack.ProtocolTypeUDP:
		return ProtocolUDP, nil
	default:
		return 0, errors.New("unknown protocol")
	}
}

// =============================================================================
// IPv4 Protocol
// =============================================================================

type IPv4 struct {
	netstack.IProtocol
	Icmp *ICMPv4
}

func NewIPv4() *IPv4 {
	ipv4 := &IPv4{
		IProtocol: netstack.NewIProtocol(netstack.ProtocolTypeIPv4),
	}
	ipv4.Log = netstack.NewLogger("IPV4")

	return ipv4
}

func (ipv4 *IPv4) HandleRx(skb *netstack.SkBuff) {
	// Create a new IPv4 header
	ipv4Header := &IPv4Header{}

	// Unmarshal the IPv4 header
	if err := ipv4Header.Unmarshal(skb.Data); err != nil {
		// If there is a problem with the IPv4 header, we may
		// need to send a ICMP error message back to the sender.
		switch err {
		case ErrInvalidIPv4Header:
			ipv4.Icmp.SendParamProblem(skb, 0)
		case ErrTTLZero:
			ipv4.Icmp.SendTimeExceeded(skb, 0)
		case ErrInvalidCheckSum:
			ipv4.Log.Println("invalid checksum")
		}

		return
	}

	rxIface, err := skb.GetRxIface()
	if err != nil {
		ipv4.Log.Println("failed to get rx iface")
		return
	}

	// Check the Destination IP matches the IP of the interface,
	// only for global unicast addresses
	if ipv4Header.DestinationIP.IsGlobalUnicast() && !rxIface.HasIPAddr(ipv4Header.DestinationIP) {
		ipv4.Log.Println("Destination IP does not match the IP of the interface")
		return
	}

	// TODO: Check fragmentation, possibly reassemble

	// Everything is good, now update the skb before passing
	// it to the transport layer or ICMP
	skb.SetSrcIP(ipv4Header.SourceIP)
	skb.SetDstIP(ipv4Header.DestinationIP)
	skb.SetType(ipv4Header.GetL4Type())
	skb.SetL3Header(ipv4Header)
	skb.StripBytes(int(ipv4Header.IHL) * 4)

	// Check if packet is ICMP
	if ipv4Header.Protocol == ProtocolICMP {
		ipv4.Icmp.HandleRx(skb)

		return
	}

	// Send the packet up the stack to the transport layer
	ipv4.RxUp(skb)
}

func (ipv4 *IPv4) HandleTx(skb *netstack.SkBuff) {
	ipv4.Log.Println("HandleTx")

	protocolType, err := GetIPProtocolType(skb)
	if err != nil {
		skb.Error(err)
		return
	}

	// Create a new IPv4 header
	ipv4Header := &IPv4Header{
		Version:        4,
		IHL:            5,
		TypeOfService:  0,
		TotalLength:    uint16(len(skb.Data) + IPv4HeaderSize),
		Identification: 0,
		Flags:          0,
		FragmentOffset: 0,
		TTL:            64,
		Protocol:       protocolType,
		HeaderChecksum: 0,
		SourceIP:       skb.GetSrcIP().To4(),
		DestinationIP:  skb.GetDstIP().To4(),
	}

	// Calculate the checksum for the IPv4 header
	ipv4Header.HeaderChecksum = netstack.Checksum(ipv4Header.Marshal())

	// Prepend the IPv4 header to the skb
	skb.SetL3Header(ipv4Header)
	skb.PrependBytes(ipv4Header.Marshal())

	// Passing to link layer, so need to set the skb type
	// to the type of the interface
	skb.SetType(netstack.ProtocolTypeEthernet)

	// Send the skb to the next layer
	ipv4.TxDown(skb)
}
