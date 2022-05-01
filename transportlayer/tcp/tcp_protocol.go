package tcp

import "github.com/mattcarp12/go-net/netstack"

type TcpProtocol struct {
	netstack.IProtocol
}

var _ netstack.Protocol = (*TcpProtocol)(nil)

func NewTCP() *TcpProtocol {
	tcp := &TcpProtocol{
		IProtocol: netstack.NewIProtocol(netstack.ProtocolTypeTCP),
	}
	return tcp
}

func (tcp *TcpProtocol) HandleRx(skb *netstack.SkBuff) {
	// Create a new TCP header

	// Unmarshal the TCP header, handle errors

	// Handle fragmentation, possibly reassemble

	// Strip the TCP header from the skb

	// Everything is good, send to user space
}

func (tcp *TcpProtocol) HandleTx(skb *netstack.SkBuff) {
	// Create new TCP header

	// Calculate checksum

	// Set the skb's L4 header to the TCP header

	// Prepend the TCP header to the skb

	// Passing to the network layer, so set the skb type
	// to the type of the socket address family (ipv4, ipv6, etc)

	// Send to the network layer
	tcp.TxDown(skb)
}
