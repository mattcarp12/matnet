package linklayer

import (
	"errors"
	"log"
	"net"

	"github.com/mattcarp12/matnet/netstack"
)

type NeighborProtocol interface {
	HandleRx(skb *netstack.SkBuff)
	Resolve(ip net.IP) (net.HardwareAddr, error)
	SendRequest(srcIP, dstIP net.IP, iface netstack.NetworkInterface)
}

type neighborSubsystem struct {
	arp *ARPProtocol
}

func NewNeighborSubsystem(arp *ARPProtocol) *neighborSubsystem {
	return &neighborSubsystem{arp}
}

var ErrProtocolNotSupported = errors.New("protocol not supported")

func (neigh *neighborSubsystem) Resolve(ip net.IP) (net.HardwareAddr, error) {
	arpLog.Println("Resolving", ip)
	if ip.To4() != nil {
		return neigh.arp.Resolve(ip)
	} else if ip.To16() != nil {
		arpLog.Printf("IPv6 not supported yet")
		return nil, ErrProtocolNotSupported
	} else {
		arpLog.Printf("Unknown IP version")
		return nil, ErrProtocolNotSupported
	}
}

// SendRequest sends an ARP request to the specified IP address
// This function should be called as a goroutine
func (neigh *neighborSubsystem) SendRequest(srcIP, dstIP net.IP, iface netstack.NetworkInterface) {
	arpLog.Println("Sending ARP request for", dstIP)
	if dstIP.To4() != nil {
		neigh.arp.ARPRequest(srcIP, dstIP, iface)
	} else if dstIP.To16() != nil {
		log.Printf("IPv6 not supported yet")
	} else {
		log.Printf("Unknown IP version")
	}
}

func (neigh *neighborSubsystem) HandleRx(skb *netstack.SkBuff) {
	arp := neigh.arp
	arp.HandleRx(skb)
}
