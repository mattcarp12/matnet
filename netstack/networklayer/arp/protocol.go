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
	if err := arpHeader.Unmarshal(skb.Data); err != nil {
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
	if !arpHeader.TargetIPAddr.Equal(skb.NetworkInterface.GetNetworkAddr()) {
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
	arpReplyHeader.SourceHWAddr = inSkb.NetworkInterface.GetHWAddr()
	arpReplyHeader.SourceIPAddr = inArpHeader.TargetIPAddr

	// Create a new arp skb
	arpReplySkb := netstack.NewSkBuff(arpReplyHeader.Marshal())

	// Set the network interface
	arpReplySkb.NetworkInterface = inSkb.NetworkInterface

	// Set L3 fields in the skb
	arpReplySkb.SrcAddr.IP = arpReplyHeader.SourceIPAddr
	arpReplySkb.DestAddr.IP = arpReplyHeader.TargetIPAddr
	arpReplySkb.L3ProtocolType = netstack.ProtocolTypeARP

	// Set the type of the skb to the link layer type (ethernet, etc),
	// which we get from the network interface
	arpReplySkb.ProtocolType = arpReplySkb.NetworkInterface.GetType()

	// Send the arp reply down to link layer
	arp.TxDown(arpReplySkb)

	// Get the skb response
	skbResp := arpReplySkb.GetResp()

	// log the response
	log.Printf("ARP Reply SkbResponse is: %+v", skbResp)
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
	arpRequestHeader.SourceIPAddr = iface.GetNetworkAddr().To4()
	arpRequestHeader.TargetIPAddr = ip.To4()
	arpRequestHeader.TargetHWAddr = net.HardwareAddr{0xff, 0xff, 0xff, 0xff, 0xff, 0xff} // Broadcast address

	// Since we're making an ARP request for the target IP, we need to
	// set the target hardware address to the broadcast address (ff:ff:ff:ff:ff:ff)
	// in the arp cache.
	arp.cache.Put(ip, net.HardwareAddr{0xff, 0xff, 0xff, 0xff, 0xff, 0xff})

	// Create a new arp skb
	rawArpHeader := arpRequestHeader.Marshal()
	arpSkb := netstack.NewSkBuff(rawArpHeader)

	// Set src and dest addresses in the skb
	arpSkb.SrcAddr.IP = arpRequestHeader.SourceIPAddr
	arpSkb.DestAddr.IP = arpRequestHeader.TargetIPAddr
	arpSkb.L3ProtocolType = netstack.ProtocolTypeARP

	// Set the network interface
	arpSkb.NetworkInterface = iface

	// Set the type of the skb to the link layer type (ethernet, etc),
	// which we get from the network interface
	arpSkb.ProtocolType = iface.GetType()

	// Send the arp request down to link layer
	arp.TxDown(arpSkb)

	// Get the skb response
	skbResp := arpSkb.GetResp()

	// log the response
	log.Printf("ARP Request SkbResponse is: %+v", skbResp)
}

// This is not used, use ARPRequest instead
func (arp *ARPProtocol) HandleTx(skb *netstack.SkBuff) {}

func (arp *ARPProtocol) Resolve(ip net.IP) (net.HardwareAddr, error) {
	return arp.cache.Lookup(ip)
}
