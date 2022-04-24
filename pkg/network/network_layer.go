package network

import (
	"log"

	"github.com/mattcarp12/go-net/pkg/link"
	"github.com/mattcarp12/go-net/pkg/protocols"
	"github.com/mattcarp12/go-net/pkg/skbuff"
)

type NetworkLayer struct {
	skbuff.SkBuffReaderWriter
	linklayer link.LinkLayer

	// Network Layer protocols
	arp  ARP
	ipv4 IPV4
	ipv6 IPV6
}

func (nl NetworkLayer) RxDispatch() {
	for {
		// Read sk_buff from network layer
		skb := <-nl.linklayer.RxChan()
		// Dispatch to appropriate protocol
		switch skb.L3_Protocol {
		case protocols.ARP:
			nl.arp.RxChan() <- skb
		case protocols.IPv4:
			nl.ipv4.RxChan() <- skb
		case protocols.IPv6:
			nl.ipv6.RxChan() <- skb
		default:
			log.Printf("Unhandled L3 protocol: %v", skb.L3_Protocol)
		}

	}
}

func Init(ll link.LinkLayer) NetworkLayer {
	nl := NetworkLayer{}
	nl.SkBuffReaderWriter = skbuff.NewSkBuffChannels()
	nl.linklayer = ll
	nl.arp = NewARP()
	nl.ipv4 = NewIPV4()
	nl.ipv6 = NewIPV6()

	// TODO: Are these started in the correct order?
	go nl.RxDispatch()
	go RxLoop(nl.arp)
	go RxLoop(nl.ipv4)
	go RxLoop(nl.ipv6)

	return nl
}

func RxLoop(protocol NetworkProtocol) {
	for {
		// Network protocol reads sk_buff from it's rx_chan
		skb := <-protocol.RxChan()

		// Handle sk_buff
		protocol.HandleRX(skb)
	}
}
