package linklayer

import (
	"net"

	"github.com/mattcarp12/go-net/pkg/entities/netdev"
	"github.com/mattcarp12/go-net/pkg/entities/protocols"
	"github.com/mattcarp12/go-net/pkg/tuntap"
)

type LinkLayer struct {
	protocols.PDUReaderWriter
	protocols.ILayer

	// TODO: support multiple devices
	tap  *TAPDevice
	loop *LoopbackDevice
}

func NewLinkLayer(tap *TAPDevice, loop *LoopbackDevice, eth *Ethernet) *LinkLayer {
	ll := &LinkLayer{}
	ll.PDUReaderWriter = protocols.NewPDUChannels()
	ll.AddProtocol(protocols.ProtocolTypeEthernet, eth)
	ll.tap = tap
	ll.loop = loop
	return ll
}

func Init() *LinkLayer {
	// Create network devices
	tap := NewTap(tuntap.TapInit("tap0", tuntap.DefaultIPv4Addr),
		net.IPv4(10, 88, 45, 69), net.HardwareAddr{0xde, 0xad, 0xbe, 0xef, 0xde, 0xad})
	loop := NewLoopback(net.IPv4(127, 0, 0, 1), net.HardwareAddr{0x00, 0x00, 0x00, 0x00, 0x00, 0x00})

	// Create L2 protocols
	eth := NewEthernet()

	// Create Link Layer
	ll := NewLinkLayer(tap, loop, eth)

	// Give Devices pointers to Link Layer
	tap.LinkLayer = ll
	loop.LinkLayer = ll

	// Give Ethernet protocol pointer to Link Layer
	eth.layer = ll

	// Start device goroutines
	go netdev.RxLoop(tap)
	go netdev.RxLoop(loop)

	// Start protocol goroutines
	ethProto, _ := ll.GetProtocol(protocols.ProtocolTypeEthernet)
	go protocols.RxLoop(ethProto)

	// Start link layer goroutines
	go protocols.RxDispatch(ll)

	return ll
}
