package networklayer

import (
	"github.com/mattcarp12/matnet/netstack"
	"github.com/mattcarp12/matnet/netstack/linklayer"
)

type NetworkLayer struct {
	netstack.ILayer
}

func Init(linkLayer *linklayer.LinkLayer) *NetworkLayer {
	networkLayer := &NetworkLayer{}
	networkLayer.SkBuffReaderWriter = netstack.NewSkBuffChannels()

	arp := NewARP()
	linkLayer.AddNeighborProtocol(arp)

	ipv4 := NewIPv4()
	icmpv4 := NewICMPv4(ipv4)
	ipv4.Icmp = icmpv4

	ipv6 := NewIPV6()

	// Add Network Layer protocols to Network Layer
	networkLayer.AddProtocol(ipv4)
	networkLayer.AddProtocol(ipv6)
	networkLayer.AddProtocol(arp)

	// Set Network Layer as the Layer for the protocols
	ipv4.SetLayer(networkLayer)
	ipv6.SetLayer(networkLayer)
	arp.SetLayer(networkLayer)

	// Set Network Layer as the next layer for Link Layer
	linkLayer.SetNextLayer(networkLayer)

	// Set Link Layer as previous layer for Network Layer
	networkLayer.SetPrevLayer(linkLayer)

	// Start protocol goroutines
	netstack.StartProtocol(ipv4)
	netstack.StartProtocol(ipv6)

	// Start network layer goroutines
	netstack.StartLayer(networkLayer)

	return networkLayer
}
