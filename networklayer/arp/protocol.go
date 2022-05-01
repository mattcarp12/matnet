package arp

import (
	"log"
	"net"

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

func (arp *ARPProtocol) HandleRx(skb *netstack.SkBuff) {
	// Create empty arp header
	arpHeader := &ARPHeader{}

	// parse the arp header
	err := arpHeader.Unmarshal(skb.GetBytes())
	if err != nil {
		log.Printf("Error parsing arp header: %v", err)
		return
	}

	log.Printf("ARP Header is: %+v", arpHeader)

	// Check if arp Hardware type is Ethernet
	if arpHeader.HardwareType != ARP_HardwareTypeEthernet {
		log.Printf("Unsupported hardware type: %v", arpHeader.HardwareType)
		return
	}

	// Check if arp protocol type is IPv4
	if arpHeader.ProtocolType != ARP_ProtocolTypeIPv4 {
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
	if !arpHeader.TargetIPAddr.Equal(skb.GetNetworkInterface().GetNetworkAddr()) {
		log.Printf("ARP request not for our IP: %v", arpHeader.TargetIPAddr)
		return
	}

	arp.ARPReply(arpHeader, skb)
}

func (arp *ARPProtocol) ARPReply(inArpHeader *ARPHeader, inSkb *netstack.SkBuff) {
	// Create a new arp header from the request header
	arpReplyHeader := &ARPHeader{}
	arpReplyHeader.HardwareType = inArpHeader.HardwareType
	arpReplyHeader.ProtocolType = inArpHeader.ProtocolType
	arpReplyHeader.HardwareSize = inArpHeader.HardwareSize
	arpReplyHeader.ProtocolSize = inArpHeader.ProtocolSize
	arpReplyHeader.OpCode = ARP_REPLY
	arpReplyHeader.TargetHWAddr = inArpHeader.SourceHWAddr
	arpReplyHeader.TargetIPAddr = inArpHeader.SourceIPAddr
	arpReplyHeader.SourceHWAddr = inSkb.GetNetworkInterface().GetHWAddr()
	arpReplyHeader.SourceIPAddr = inArpHeader.TargetIPAddr

	log.Printf("ARP Reply Header is: %+v", arpReplyHeader)

	// Create a new arp skb
	rawArpHeader, err := arpReplyHeader.Marshal()
	if err != nil {
		log.Printf("Error marshaling arp header: %v", err)
		return
	}
	arpReplySkb := netstack.NewSkBuff(rawArpHeader)

	// Set the network interface
	arpReplySkb.SetNetworkInterface(inSkb.GetNetworkInterface())

	// Set the skb L3 header to the ARP header
	arpReplySkb.SetL3Header(arpReplyHeader)

	// Set the type of the skb to the link layer type (ethernet, etc),
	// which we get from the network interface
	arpReplySkb.SetType(arpReplySkb.GetNetworkInterface().GetType())

	// Send the arp reply down to link layer
	arp.TxDown(arpReplySkb)
}

func (arp *ARPProtocol) ARPRequest(ip net.IP, iface netstack.NetworkInterface) {
	// Create a new arp header
	arpRequestHeader := &ARPHeader{}
	arpRequestHeader.HardwareType = ARP_HardwareTypeEthernet
	arpRequestHeader.ProtocolType = ARP_ProtocolTypeIPv4
	arpRequestHeader.HardwareSize = 6
	arpRequestHeader.ProtocolSize = 4
	arpRequestHeader.OpCode = ARP_REQUEST
	arpRequestHeader.SourceHWAddr = iface.GetHWAddr()
	arpRequestHeader.SourceIPAddr = iface.GetNetworkAddr()
	arpRequestHeader.TargetIPAddr = ip
	arpRequestHeader.TargetHWAddr = net.HardwareAddr{0xff, 0xff, 0xff, 0xff, 0xff, 0xff} // Broadcast address

	// Create a new arp skb
	rawArpHeader, err := arpRequestHeader.Marshal()
	if err != nil {
		log.Printf("Error marshaling arp header: %v", err)
		return
	}
	arpSkb := netstack.NewSkBuff(rawArpHeader)

	// Set L3 header to the arp header
	arpSkb.SetL3Header(arpRequestHeader)

	// Set the network interface
	arpSkb.SetNetworkInterface(iface)

	// Set the type of the skb to the link layer type (ethernet, etc),
	// which we get from the network interface
	arpSkb.SetType(iface.GetType())

	// Send the arp request down to link layer
	arp.TxDown(arpSkb)
}

// This is not used, use ARPRequest instead
func (arp *ARPProtocol) HandleTx(skb *netstack.SkBuff) {}

func (arp *ARPProtocol) Resolve(ip net.IP) (net.HardwareAddr, error) {
	return arp.cache.Lookup(ip)
}
