package arp

import (
	"log"

	"github.com/mattcarp12/go-net/netstack"
)

type ARPProtocol struct {
	netstack.IProtocol
	cache *ARPCache
}

func NewARP() *ARPProtocol {
	return &ARPProtocol{
		IProtocol: netstack.NewIProtocol(netstack.ProtocolTypeARP),
		cache:     NewARPCache(),
	}
}

func (ap *ARPProtocol) GetType() netstack.ProtocolType {
	return netstack.ProtocolTypeARP
}

func (arp *ARPProtocol) HandleRx(skb *netstack.SkBuff) {
	// parse the arp header
	arpHeader, err := ParseARPHeader(skb.GetBytes())
	if err != nil {
		log.Printf("Error parsing arp header: %v", err)
		return
	}

	// Check if arp Hardware type is Ethernet
	if arpHeader.HardwareType != HardwareTypeEthernet {
		log.Printf("Unsupported hardware type: %v", arpHeader.HardwareType)
		return
	}

	// Check if arp protocol type is IPv4
	if arpHeader.ProtocolType != ProtocolTypeIPv4 {
		log.Printf("Unsupported protocol type: %v", arpHeader.ProtocolType)
		return
	}

	// Update the arp cache with the entry for source ip
	arp.cache.Update(arpHeader)

	// Check if this is an arp request
	if arpHeader.OpCode != ARP_REQUEST {
		return
	}

	// Make sure TargetIP equals our IP
	// TODO: Handle multihoming? Handle ipv4/ipv6 addresses?
	if arpHeader.TargetIPAddr.Equal(skb.GetNetworkInterface().GetNetworkAddr()) {
		return
	}

	arp.ARPReply(arpHeader, skb)
}

func (ap *ARPProtocol) ARPReply(inArpHeader *ARPHeader, inSkb *netstack.SkBuff) {
	// Create a new arp header from the request header
	arpReplyHeader := ARPHeader{}
	arpReplyHeader.HardwareType = inArpHeader.HardwareType
	arpReplyHeader.ProtocolType = inArpHeader.ProtocolType
	arpReplyHeader.HardwareSize = inArpHeader.HardwareSize
	arpReplyHeader.ProtocolSize = inArpHeader.ProtocolSize
	arpReplyHeader.OpCode = ARP_REPLY
	arpReplyHeader.TargetHWAddr = inArpHeader.SourceHWAddr
	arpReplyHeader.TargetIPAddr = inArpHeader.SourceIPAddr
	arpReplyHeader.SourceHWAddr = inSkb.GetNetworkInterface().GetHWAddr()
	arpReplyHeader.SourceIPAddr = inArpHeader.TargetIPAddr

	// Create a new arp skb
	arpReplySkb := netstack.NewSkBuff(arpReplyHeader.Marshal())

	// Set the network interface
	arpReplySkb.SetNetworkInterface(inSkb.GetNetworkInterface())

	// The "L3Header" is just the ARP payload, so don't need to set it

	// Set the type of the skb to the link layer type (ethernet, etc),
	// which we get from the network interface
	arpReplySkb.SetType(arpReplySkb.GetNetworkInterface().GetType())

	// Send the arp reply
	ap.GetLayer().GetPrevLayer().TxChan() <- arpReplySkb
}

func (arp *ARPProtocol) HandleTx(skb *netstack.SkBuff) {
	log.Printf("ARP: %v", skb)
}
