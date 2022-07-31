package networklayer

import (
	"github.com/mattcarp12/matnet/netstack"
	"github.com/mattcarp12/matnet/netstack/linklayer"
)

func Init(linkLayer *linklayer.LinkLayer) *netstack.Layer {
	arp := NewARP()
	linkLayer.AddNeighborProtocol(arp)

	ipv4 := NewIPv4()
	icmpv4 := NewICMPv4(ipv4)
	ipv4.Icmp = icmpv4

	ipv6 := NewIPV6()

	networkLayer := netstack.NewLayer(ipv4, ipv6, arp)

	// Set Network Layer as the Layer for the protocols
	ipv4.SetLayer(networkLayer)
	ipv6.SetLayer(networkLayer)
	arp.SetLayer(networkLayer)

	// Set Network Layer as the next layer for Link Layer
	linkLayer.SetNextLayer(networkLayer)

	// Set Link Layer as previous layer for Network Layer
	networkLayer.SetPrevLayer(linkLayer.Layer)

	// Start protocol goroutines
	netstack.StartProtocol(ipv4)
	netstack.StartProtocol(ipv6)

	// Start network layer goroutines
	networkLayer.StartLayer()

	return networkLayer
}
