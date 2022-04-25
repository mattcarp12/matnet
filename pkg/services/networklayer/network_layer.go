package networklayer

import (
	"github.com/mattcarp12/go-net/pkg/entities/protocols"
	"github.com/mattcarp12/go-net/pkg/services/linklayer"
)

type NetworkLayer struct {
	protocols.PDUReaderWriter
	protocols.ILayer
}

func Init(ll *linklayer.LinkLayer) *NetworkLayer {
	nl := &NetworkLayer{}
	nl.PDUReaderWriter = protocols.NewPDUChannels()

	// Create Network Layer protocols
	arp := NewARP()
	ipv4 := NewIPV4()
	ipv6 := NewIPV6()

	// Add Network Layer protocols to Network Layer
	nl.AddProtocol(protocols.ProtocolTypeARP, arp)
	nl.AddProtocol(protocols.ProtocolTypeIPv4, ipv4)
	nl.AddProtocol(protocols.ProtocolTypeIPv6, ipv6)

	// Set Network Layer as the next layer for Link Layer
	ll.SetNextLayer(nl)

	// Set Link Layer as previous layer for Network Layer
	nl.SetPrevLayer(ll)

	go protocols.RxDispatch(nl)
	go protocols.RxLoop(arp)
	go protocols.RxLoop(ipv4)
	go protocols.RxLoop(ipv6)

	return nl
}
