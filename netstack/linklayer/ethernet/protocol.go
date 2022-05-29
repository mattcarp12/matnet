package ethernet

import (
	logging "log"
	"os"

	"github.com/mattcarp12/go-net/netstack"
)

var log = logging.New(os.Stdout, "[Ethernet] ", logging.Ldate|logging.Lmicroseconds|logging.Lshortfile)

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
	if err := eth_header.Unmarshal(skb.GetBytes()); err != nil {
		log.Printf("Error parsing ethernet header: %v", err)
		return
	}

	// Check if unicast
	if IsUnicast(eth_header.addr.DstAddr) {
		// Check if the packed is destined for this interface
		if eth_header.addr.DstAddr.String() != skb.GetNetworkInterface().GetHWAddr().String() {
			log.Printf("Packet not for this interface (dst: %s, src: %s)", eth_header.addr.DstAddr.String(), skb.GetNetworkInterface().GetHWAddr().String())
			return
		}
	} // If multicast or broadcast, continue processing

	// Set link layer header
	skb.SetL2Header(&eth_header)

	// Set skb type to the next layer type (ipv4, arp, etc)
	// by inspecting EtherType field
	skb.SetType(eth_header.GetL3Type())

	// Strip ethernet header from skb data buffer
	skb.StripBytes(EthernetHeaderSize)

	// Pass to network layer
	eth.GetLayer().GetNextLayer().RxChan() <- skb
}

// HandleTx is called when a skbuff from the network layer is ready to be sent
func (eth *Ethernet) HandleTx(skb *netstack.SkBuff) {
	log.Println("Ethernet: HandleTx")
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
		go eth.arp.SendRequest(skb.GetL3Header().GetDstIP(), skb.GetNetworkInterface())

		// Make sure to send response for the dropped skb
		// TODO: Make request cache to handle requests once the arp response is received
		skb.Error(err)

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

	log.Printf("Ethernet: Sending packet to %s", destHWAddr.String())
	log.Printf("Ethernet Header: %+v", eth_header)

	// Prepend ethernet header to skbuff
	skb.PrependBytes(eth_header.Marshal())

	// Pass to network interface
	skb.GetNetworkInterface().TxChan() <- skb
}
