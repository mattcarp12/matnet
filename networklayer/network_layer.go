package networklayer

import (
	"github.com/mattcarp12/go-net/networklayer/arp"
	"github.com/mattcarp12/go-net/networklayer/ipv4"
	"github.com/mattcarp12/go-net/networklayer/ipv6"
	"github.com/mattcarp12/go-net/netstack"
	"github.com/mattcarp12/go-net/linklayer"
)

type NetworkLayer struct {
	netstack.SkBuffReaderWriter
	netstack.ILayer
}

func Init(link_layer *linklayer.LinkLayer) *NetworkLayer {
	network_layer := &NetworkLayer{}
	network_layer.SkBuffReaderWriter = netstack.NewSkBuffChannels()

	// Create Network Layer protocols
	arp := arp.NewARP()
	ipv4 := ipv4.NewIPV4()
	ipv6 := ipv6.NewIPV6()

	// Add Network Layer protocols to Network Layer
	network_layer.AddProtocol(arp)
	network_layer.AddProtocol(ipv4)
	network_layer.AddProtocol(ipv6)

	// Set Network Layer as the Layer for the protocols
	arp.SetLayer(network_layer)
	ipv4.SetLayer(network_layer)
	ipv6.SetLayer(network_layer)

	// Set Network Layer as the next layer for Link Layer
	link_layer.SetNextLayer(network_layer)

	// Set Link Layer as previous layer for Network Layer
	network_layer.SetPrevLayer(link_layer)

	// Create neighbor resolution subsystem
	neigh := &neighborSubsystem{arp}
	link_layer.SetNeighborProtocol(neigh)

	go netstack.RxDispatch(network_layer)
	go netstack.ProtocolRxLoop(arp)
	go netstack.ProtocolRxLoop(ipv4)
	go netstack.ProtocolRxLoop(ipv6)

	return network_layer
}
