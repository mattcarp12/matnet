package linklayer

import (
	"net"

	"github.com/mattcarp12/matnet/netstack"
	"github.com/mattcarp12/matnet/tuntap"
)

type LinkLayer struct {
	netstack.ILayer
	tap  *TAPDevice
	loop *LoopbackDevice
}

func NewLinkLayer(tap *TAPDevice, loop *LoopbackDevice, eth *EthernetProtocol) *LinkLayer {
	ll := &LinkLayer{}
	ll.SkBuffReaderWriter = netstack.NewSkBuffChannels()
	ll.AddProtocol(eth)
	ll.tap = tap
	ll.loop = loop
	return ll
}

func (ll *LinkLayer) SetNeighborProtocol(neigh NeighborProtocol) {
	eth, err := ll.GetProtocol(netstack.ProtocolTypeEthernet)
	if err != nil {
		panic(err)
	}
	eth.(*EthernetProtocol).SetNeighborProtocol(neigh)
}

const DefaultIPAddr = "10.88.45.69"
const DefaultGateway = "10.88.45.1"

func Init() (*LinkLayer, netstack.RoutingTable) {
	// Create network devices
	tap := NewTap(
		tuntap.TapInit("tap0", tuntap.DefaultIPv4Addr),
		"tap0",
		net.HardwareAddr{0xde, 0xad, 0xbe, 0xef, 0xde, 0xad},
		[]netstack.IfAddr{
			{
				IP:      net.ParseIP(DefaultIPAddr),
				Netmask: net.IPv4Mask(255, 255, 255, 0),
				Gateway: net.ParseIP(DefaultGateway),
			},
		},
	)

	loop := NewLoopback()

	// Create L2 protocols
	eth := NewEthernet()

	// Create Link Layer
	link_layer := NewLinkLayer(tap, loop, eth)

	arp := NewARP()
	arp.SetLayer(link_layer)

	neigh := NewNeighborSubsystem(arp)

	link_layer.SetNeighborProtocol(neigh)

	// Give Devices pointers to Link Layer
	tap.LinkLayer = link_layer
	loop.LinkLayer = link_layer

	// Give Ethernet protocol pointer to Link Layer
	eth.SetLayer(link_layer)

	// Start device goroutines
	netstack.StartInterface(tap)
	netstack.StartInterface(loop)

	// Start protocol goroutines
	netstack.StartProtocol(eth)

	// Start link layer goroutines
	netstack.StartLayer(link_layer)

	// Make routing table
	rt := netstack.NewRoutingTable()
	rt.AddConnectedRoutes(tap)
	rt.SetDefaultRoute(
		net.IPNet{
			IP:   net.ParseIP(DefaultIPAddr),
			Mask: net.IPv4Mask(255, 255, 255, 0),
		},
		net.ParseIP(DefaultGateway),
		tap,
	)
	rt.AddConnectedRoutes(loop)

	return link_layer, rt
}
