package ipv4

import (
	"log"

	"github.com/mattcarp12/go-net/netstack"
)

type IPv4 struct {
	netstack.IProtocol
	icmp *ICMPv4
}

func NewIPv4() *IPv4 {
	return &IPv4{
		IProtocol: netstack.NewIProtocol(netstack.ProtocolTypeIPv4),
	}
}

func (ipv4 *IPv4) HandleRx(skb *netstack.SkBuff) {
	// Create a new IPv4 header
	ipv4Header := &IPv4Header{}

	// Unmarshal the IPv4 header, handle errors
	err := ipv4Header.Unmarshal(skb.Data)
	if err != nil {
		switch err {
		case ErrInvalidIPv4Header:
			ipv4.icmp.SendParamProblem(skb, 0)
		case ErrTTLZero:
			ipv4.icmp.SendTimeExceeded(skb, 0)
		case ErrInvalidCheckSum:
			log.Println("invalid checksum")
		}
		return
	}

	// Check the Destination IP matches the IP of the interface
	// TODO: Handle dual stack and multihoming
	if ipv4Header.DestinationIP.IsGlobalUnicast() {
		// Only handle global unicast addresses
		if ipv4Header.DestinationIP.String() != skb.GetNetworkInterface().GetNetworkAddr().String() {
			log.Println("Destination IP does not match the IP of the interface")
			return
		}
	}

	// TODO: Check fragmentation, possibly reassemble

	// Everything is good, now update the skb before passing
	// it to the transport layer or ICMP
	skb.SetL3Header(ipv4Header)
	skb.StripBytes(int(ipv4Header.IHL) * 4)
	skb.SetType(ipv4Header.GetL4Type())

	// Check if packet is ICMP
	if ipv4Header.Protocol == ProtocolICMP {
		ipv4.icmp.HandleRx(skb)
		return
	}

	// Check if packet is TCP or UDP
	// If so, pass it up to Layer 4
	if ipv4Header.Protocol == ProtocolTCP || ipv4Header.Protocol == ProtocolUDP {
		// TODO: implement this
		// ipv4.RxUp(skb)
		return
	}

	// If packet type is not recognized, drop it
	log.Printf("IPV4: Unknown packet type: %v", ipv4Header.Protocol)
}

func (ipv4 *IPv4) HandleTx(skb *netstack.SkBuff) {
	// Get destination address from upper layer header
	sock := skb.GetSocket()
	if sock == nil {
		log.Println("IPV4: Socket is nil")
		return
	}
	sockAddr := sock.GetAddress()
	destIP := sockAddr.DestIP

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
		Protocol:       GetIPProtocolType(skb.GetType()),
		HeaderChecksum: 0,
		SourceIP:       skb.GetNetworkInterface().GetNetworkAddr(),
		DestinationIP:  destIP,
	}

	// Calculate the checksum
	rawHeader, _ := ipv4Header.Marshal()
	ipv4Header.HeaderChecksum = Checksum(rawHeader)

	// Set the skb's L3 header
	skb.SetL3Header(ipv4Header)

	// Marshal the IPv4 header
	ipv4HeaderBytes, _ := ipv4Header.Marshal()

	// Prepend the IPv4 header to the skb
	skb.PrependBytes(ipv4HeaderBytes)

	// Passing to link layer, so need to set the skb type
	// to the type of the interface
	skb.SetType(skb.GetNetworkInterface().GetType())

	// Send the skb to the next layer
	ipv4.TxDown(skb)
}

func (ipv4 *IPv4) SetIcmp(icmp *ICMPv4) {
	ipv4.icmp = icmp
}
