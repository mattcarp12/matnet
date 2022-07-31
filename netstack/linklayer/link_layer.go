package linklayer

import (
	"net"

	"github.com/mattcarp12/matnet/netstack"
	"github.com/mattcarp12/matnet/tuntap"
)

type LinkLayer struct {
	*netstack.Layer
	tap  *TAPDevice
	loop *LoopbackDevice
}

func (ll *LinkLayer) SetNeighborSubsystem(neigh *NeighborSubsystem) {
	eth, err := ll.GetProtocol(netstack.ProtocolTypeEthernet)
	if err != nil {
		panic(err)
	}

	eth.(*EthernetProtocol).SetNeighborSubsystem(neigh)
}

func (ll *LinkLayer) AddNeighborProtocol(prot NeighborProtocol) {
	eth, err := ll.GetProtocol(netstack.ProtocolTypeEthernet)
	if err != nil {
		panic(err)
	}

	eth.(*EthernetProtocol).AddNeighborProtocol(prot)
}

func Init() (*LinkLayer, netstack.RoutingTable) {
	// Create network devices
	tapMAC, err := net.ParseMAC(netstack.DefaultMACAddr)
	if err != nil {
		panic(err)
	}

	tap := NewTap(
		tuntap.TapInit("tap0", tuntap.DefaultIPv4Addr),
		"tap0",
		tapMAC,
		[]netstack.IfAddr{
			{
				IP:      net.ParseIP(netstack.DefaultIPAddr),
				Netmask: net.IPv4Mask(255, 255, 255, 0),
				Gateway: net.ParseIP(netstack.DefaultGateway),
			},
		},
	)

	loop := NewLoopback()

	// Create L2 protocols
	eth := NewEthernet()

	// Create Link Layer
	linkLayer := &LinkLayer{
		Layer: netstack.NewLayer(eth),
		tap:   tap,
		loop:  loop,
	}

	neigh := NewNeighborSubsystem()
	linkLayer.SetNeighborSubsystem(neigh)

	// Give Devices pointers to Link Layer
	tap.LinkLayer = linkLayer
	loop.LinkLayer = linkLayer

	// Give Ethernet protocol pointer to Link Layer
	eth.SetLayer(linkLayer.Layer)

	// Start device goroutines
	netstack.StartInterface(tap)
	netstack.StartInterface(loop)

	// Start protocol goroutines
	netstack.StartProtocol(eth)

	// Start link layer goroutines
	linkLayer.StartLayer()

	// Make routing table
	routingTable := netstack.NewRoutingTable()
	routingTable.AddConnectedRoutes(tap)
	routingTable.SetDefaultRoute(
		net.IPNet{
			IP:   net.ParseIP(netstack.DefaultIPAddr),
			Mask: net.IPv4Mask(255, 255, 255, 0),
		},
		net.ParseIP(netstack.DefaultGateway),
		tap,
	)
	routingTable.AddConnectedRoutes(loop)

	return linkLayer, routingTable
}
