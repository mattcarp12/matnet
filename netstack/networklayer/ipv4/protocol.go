package ipv4

import (
	"errors"
	"log"

	"github.com/mattcarp12/go-net/netstack"
)

type IPv4 struct {
	netstack.IProtocol
	Icmp *ICMPv4
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
			ipv4.Icmp.SendParamProblem(skb, 0)
		case ErrTTLZero:
			ipv4.Icmp.SendTimeExceeded(skb, 0)
		case ErrInvalidCheckSum:
			log.Println("invalid checksum")
		}
		return
	}

	// Check the Destination IP matches the IP of the interface
	// TODO: Handle dual stack and multihoming
	if ipv4Header.DestinationIP.IsGlobalUnicast() {
		// Only handle global unicast addresses
		if ipv4Header.DestinationIP.String() != skb.NetworkInterface.GetNetworkAddr().String() {
			log.Println("Destination IP does not match the IP of the interface")
			return
		}
	}

	// TODO: Check fragmentation, possibly reassemble

	// Everything is good, now update the skb before passing
	// it to the transport layer or ICMP
	skb.SrcAddr.IP = ipv4Header.SourceIP
	skb.DestAddr.IP = ipv4Header.DestinationIP
	skb.ProtocolType = ipv4Header.GetL4Type()
	skb.StripBytes(int(ipv4Header.IHL) * 4)

	// Check if packet is ICMP
	if ipv4Header.Protocol == ProtocolICMP {
		ipv4.Icmp.HandleRx(skb)
		return
	}

	// Check if packet is TCP or UDP
	// If so, pass it up to Layer 4
	if ipv4Header.Protocol == ProtocolTCP || ipv4Header.Protocol == ProtocolUDP {
		ipv4.RxUp(skb)
		return
	}

	// If packet type is not recognized, drop it
	log.Printf("IPV4: Unknown packet type: %v", ipv4Header.Protocol)
	skb.Error(errors.New("unknown packet type"))
}

func (ipv4 *IPv4) HandleTx(skb *netstack.SkBuff) {
	log.Println("IPV4: HandleTx")

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
		SourceIP:       skb.SrcAddr.IP,
		DestinationIP:  skb.DestAddr.IP,
	}

	// Calculate the checksum
	ipv4Header.HeaderChecksum = netstack.Checksum(ipv4Header.Marshal())

	// Prepend the IPv4 header to the skb
	skb.PrependBytes(ipv4Header.Marshal())

	// Passing to link layer, so need to set the skb type
	// to the type of the interface
	skb.ProtocolType = netstack.ProtocolTypeEthernet

	// Send the skb to the next layer
	ipv4.TxDown(skb)
}
