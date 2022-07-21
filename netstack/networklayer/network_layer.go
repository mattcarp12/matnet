package networklayer

import (
	"github.com/mattcarp12/matnet/netstack"
	"github.com/mattcarp12/matnet/netstack/linklayer"
)

type NetworkLayer struct {
	netstack.ILayer
}

func Init(link_layer *linklayer.LinkLayer) *NetworkLayer {
	network_layer := &NetworkLayer{}
	network_layer.SkBuffReaderWriter = netstack.NewSkBuffChannels()

	ipv4proto := NewIPv4()
	icmpv4 := NewICMPv4(ipv4proto)
	ipv4proto.Icmp = icmpv4

	ipv6 := NewIPV6()

	// Add Network Layer protocols to Network Layer
	network_layer.AddProtocol(ipv4proto)
	network_layer.AddProtocol(ipv6)

	// Set Network Layer as the Layer for the protocols
	ipv4proto.SetLayer(network_layer)
	ipv6.SetLayer(network_layer)

	// Set Network Layer as the next layer for Link Layer
	link_layer.SetNextLayer(network_layer)

	// Set Link Layer as previous layer for Network Layer
	network_layer.SetPrevLayer(link_layer)

	// Start protocol goroutines
	netstack.StartProtocol(ipv4proto)
	netstack.StartProtocol(ipv6)

	// Start network layer goroutines
	netstack.StartLayer(network_layer)

	return network_layer
}
