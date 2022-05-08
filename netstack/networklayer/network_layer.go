package networklayer

import (
	"github.com/mattcarp12/go-net/netstack"
	"github.com/mattcarp12/go-net/netstack/linklayer"
	"github.com/mattcarp12/go-net/netstack/networklayer/arp"
	"github.com/mattcarp12/go-net/netstack/networklayer/ipv4"
	"github.com/mattcarp12/go-net/netstack/networklayer/ipv6"
)

type NetworkLayer struct {
	netstack.ILayer
}

func Init(link_layer *linklayer.LinkLayer) *NetworkLayer {
	network_layer := &NetworkLayer{}
	network_layer.SkBuffReaderWriter = netstack.NewSkBuffChannels()

	// Create Network Layer protocols
	arp := arp.NewARP()

	ipv4proto := ipv4.NewIPv4()
	icmpv4 := ipv4.NewICMPv4(ipv4proto)
	ipv4proto.SetIcmp(icmpv4)

	ipv6 := ipv6.NewIPV6()

	// Add Network Layer protocols to Network Layer
	network_layer.AddProtocol(arp)
	network_layer.AddProtocol(ipv4proto)
	network_layer.AddProtocol(ipv6)

	// Set Network Layer as the Layer for the protocols
	arp.SetLayer(network_layer)
	ipv4proto.SetLayer(network_layer)
	ipv6.SetLayer(network_layer)

	// Set Network Layer as the next layer for Link Layer
	link_layer.SetNextLayer(network_layer)

	// Set Link Layer as previous layer for Network Layer
	network_layer.SetPrevLayer(link_layer)

	// Create neighbor resolution subsystem
	neigh := &neighborSubsystem{arp}
	link_layer.SetNeighborProtocol(neigh)

	// Start protocol goroutines
	netstack.StartProtocol(arp)
	netstack.StartProtocol(ipv4proto)
	netstack.StartProtocol(ipv6)

	// Start network layer goroutines
	netstack.StartLayerDispatchLoops(network_layer)

	return network_layer
}
