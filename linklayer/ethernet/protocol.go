package ethernet

import (
	"log"

	"github.com/mattcarp12/go-net/netstack"
)

type Ethernet struct {
	netstack.IProtocol
}

func NewEthernet() *Ethernet {
	iprotocol := netstack.NewIProtocol(netstack.ProtocolTypeEthernet)
	return &Ethernet{iprotocol}
}

func (eth *Ethernet) HandleRx(skb *netstack.SkBuff) {
	// Create empty ethernet header
	eth_header := EthernetHeader{}

	// Parse the ethernet header
	err := eth_header.Unmarshal(skb.GetBytes())
	if err != nil {
		log.Printf("Error parsing ethernet header: %v", err)
	}

	// Set skb type to the next layer type (ipv4, arp, etc)
	nextLayerType, err := eth_header.GetNextLayerType()
	if err != nil {
		log.Printf("Error getting next layer type: %v", err)
	}
	skb.SetType(nextLayerType)

	// Strip ethernet header from skb data buffer
	skb.StripBytes(EthernetHeaderSize)

	// Pass to next layer
	eth.GetLayer().GetNextLayer().RxChan() <- skb
}

func (eth *Ethernet) HandleTx(skb *netstack.SkBuff) {
	// TODO: Implement
}
