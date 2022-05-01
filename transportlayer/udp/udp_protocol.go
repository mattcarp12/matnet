package udp

import "github.com/mattcarp12/go-net/netstack"

type UdpProtocol struct {
	netstack.IProtocol
}

func NewUDP() *UdpProtocol {
	udp := &UdpProtocol{
		IProtocol: netstack.NewIProtocol(netstack.ProtocolTypeUDP),
	}
	return udp
}

func (udp *UdpProtocol) HandleRx(skb *netstack.SkBuff) {
	// Create a new UDP header
	// Unmarshal the UDP header, handle errors
	// Handle fragmentation, possibly reassemble
	// Strip the UDP header from the skb
	// Everything is good, send to user space
}

func (udp *UdpProtocol) HandleTx(skb *netstack.SkBuff) {
	// Create new UDP header
	// Calculate checksum
	// Set the skb's L4 header to the UDP header
	// Prepend the UDP header to the skb
	// Passing to the network layer, so set the skb type
	// to the type of the socket address family (ipv4, ipv6, etc)
	// Send to the network layer
	udp.TxDown(skb)
}
