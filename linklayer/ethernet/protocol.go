package ethernet

import (
	"log"

	"github.com/mattcarp12/go-net/netstack"
)

type Ethernet struct {
	netstack.IProtocol

	// arp represents the neighbor subsystem for either ipv4 (arp) or ipv6 (ndp)
	arp netstack.NeighborProtocol
}

func NewEthernet() *Ethernet {
	eth := &Ethernet{}
	eth.IProtocol = netstack.NewIProtocol(netstack.ProtocolTypeEthernet)
	return eth
}

func (eth *Ethernet) SetNeighborProtocol(arp netstack.NeighborProtocol) {
	eth.arp = arp
}

// HandleRx is called when a skbuff from the network interface is ready
// to be processed by the network stack
func (eth *Ethernet) HandleRx(skb *netstack.SkBuff) {
	// Create empty ethernet header
	eth_header := EthernetHeader{}

	// Parse the ethernet header
	err := eth_header.Unmarshal(skb.GetBytes())
	if err != nil {
		log.Printf("Error parsing ethernet header: %v", err)
	}

	// Set skb type to the next layer type (ipv4, arp, etc)
	// by inspecting EtherType field
	nextLayerType, err := eth_header.GetNextLayerType()
	if err != nil {
		log.Printf("Error getting next layer type: %v", err)
	}

	skb.SetType(nextLayerType)

	// Strip ethernet header from skb data buffer
	skb.StripBytes(EthernetHeaderSize)

	// Pass to network layer
	eth.GetLayer().GetNextLayer().RxChan() <- skb
}

// HandleTx is called when a skbuff from the network layer is ready to be sent
func (eth *Ethernet) HandleTx(skb *netstack.SkBuff) {
	/*
		If link layer header is already set, there is nothing to do
		so just pass the skbuff to network interface
	*/
	if skb.GetL2Header() != nil {
		skb.GetNetworkInterface().TxChan() <- skb
		return
	}

	/*
		If link layer header is not set, we need to create an ethernet header
		then append it to the skbuff. Then pass the skbuff to network interface.

		In order to fill out the ethernet header, we need to get the
		destination hardware address. We can get this from the arp cache.

		If the arp cache does not have the destination hardware address,
		we need to send an arp request to get it.
	*/

	// Get the destination hardware address from the arp cache
	destHWAddr, err := eth.arp.Resolve(skb.GetL3Header().GetDstIP())
	if err != nil {
		log.Printf("Error resolving destination hardware address: %v", err)
		// If we can't get the hardware address, send an arp request
		eth.arp.SendRequest(skb.GetL3Header().GetDstIP(), skb.GetNetworkInterface())
		return
	}

	// Create ethernet header
	eth_header := EthernetHeader{}
	eth_header.addr.DstAddr = destHWAddr
	eth_header.addr.SrcAddr = skb.GetNetworkInterface().GetHWAddr()
	eth_header.EtherType, err = GetEtherTypeFromProtocolType(skb.GetL3Header().GetType())
	if err != nil {
		log.Printf("Error getting EtherType: %v", err)
		return
	}

	// Prepend ethernet header to skbuff
	eth_header_bytes, err := eth_header.Marshal()
	if err != nil {
		log.Printf("Error marshaling ethernet header: %v", err)
		return
	}

	skb.PrependBytes(eth_header_bytes)

	// Pass to network interface
	iface := skb.GetNetworkInterface()
	iface.TxChan() <- skb
}
