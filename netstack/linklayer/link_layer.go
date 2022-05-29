package linklayer

import (
	logging "log"
	"net"
	"os"

	"github.com/mattcarp12/go-net/netstack"
	"github.com/mattcarp12/go-net/netstack/linklayer/ethernet"
	"github.com/mattcarp12/go-net/tuntap"
)

var log = logging.New(os.Stdout, "[LinkLayer] ", logging.Ldate|logging.Lmicroseconds|logging.Lshortfile)

type LinkLayer struct {
	netstack.ILayer

	// TODO: support multiple devices
	tap  *TAPDevice
	loop *LoopbackDevice
}

func NewLinkLayer(tap *TAPDevice, loop *LoopbackDevice, eth *ethernet.Ethernet) *LinkLayer {
	ll := &LinkLayer{}
	ll.SkBuffReaderWriter = netstack.NewSkBuffChannels()
	ll.AddProtocol(eth)
	ll.tap = tap
	ll.loop = loop
	return ll
}

func (ll *LinkLayer) SetNeighborProtocol(neigh netstack.NeighborProtocol) {
	eth, err := ll.GetProtocol(netstack.ProtocolTypeEthernet)
	if err != nil {
		panic(err)
	}
	eth.(*ethernet.Ethernet).SetNeighborProtocol(neigh)
}

func Init() (*LinkLayer, netstack.RoutingTable) {
	// Create network devices
	tapConfig := IfaceConfig{
		Name:    "tap0",
		IP:      net.IPv4(10, 88, 45, 69),
		Netmask: net.IPv4Mask(255, 255, 255, 0),
		Gateway: net.IPv4(10, 88, 45, 1),
		MAC:     net.HardwareAddr{0xde, 0xad, 0xbe, 0xef, 0xde, 0xad},
		Mtu:     1500,
	}

	tap := NewTap(tuntap.TapInit("tap0", tuntap.DefaultIPv4Addr), tapConfig)
	loop := NewLoopback()

	// Create L2 protocols
	eth := ethernet.NewEthernet()

	// Create Link Layer
	link_layer := NewLinkLayer(tap, loop, eth)

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
	netstack.StartLayerDispatchLoops(link_layer)

	// Make routing table
	rt := netstack.NewRoutingTable()
	tap_route := rt.AddConnectedRoute(tap)
	rt.SetDefaultRoute(*tap_route)
	rt.AddConnectedRoute(loop)

	return link_layer, rt
}
