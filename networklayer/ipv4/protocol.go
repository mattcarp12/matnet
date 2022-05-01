package ipv4

import (
	"log"

	"github.com/mattcarp12/go-net/netstack"
)

type IPV4 struct {
	netstack.IProtocol
	icmp *ICMP
}

func NewIPV4() *IPV4 {
	return &IPV4{
		IProtocol: netstack.NewIProtocol(netstack.ProtocolTypeIPv4),
	}
}

func (ipv4 *IPV4) HandleRx(skb *netstack.SkBuff) {
	// Create a new IPv4 header
	ipv4Header := &IPv4Header{}

	// Unmarshal the IPv4 header
	err := ipv4Header.Unmarshal(skb.Data)
	if err != nil {
		log.Println("Unmarshal error:", err)
		return
	}

	// Check if the Version is 4
	if ipv4Header.Version != 4 {
		log.Println("Invalid IPv4 header")
		return
	}

	// Check if the IHL is 5
	if ipv4Header.IHL != 5 {
		log.Println("Invalid IPv4 header")
		return
	}

	// Check the TTL
	if ipv4Header.TTL < 1 {
		log.Println("Time to live is 0")
		// TODO: Send ICMP time exceeded
		return
	}

	// Check the checksum
	if Checksum(skb.Data) != 0 {
		log.Println("Invalid checksum")
		return
	}

	// TODO: Check fragmentation, possibly reassemble

	// Strip the IPv4 header off the skb
	skb.StripBytes(int(ipv4Header.IHL) * 4)

	// Check if packet is ICMP
	if ipv4Header.Protocol == ProtocolICMP {


}

func (ipv4 *IPV4) HandleTx(skb *netstack.SkBuff) {
	// log.Printf("IPV4: %v", skb)
}
