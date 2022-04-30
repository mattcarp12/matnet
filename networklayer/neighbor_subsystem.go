package networklayer

import (
	"errors"
	"log"
	"net"

	"github.com/mattcarp12/go-net/netstack"
	"github.com/mattcarp12/go-net/networklayer/arp"
)

type neighborSubsystem struct {
	arp *arp.ARPProtocol
}

var ErrProtocolNotSupported = errors.New("Protocol not supported")

func (neigh *neighborSubsystem) Resolve(ip net.IP) (net.HardwareAddr, error) {
	if ip.To4() != nil {
		return neigh.arp.Resolve(ip)
	} else if ip.To16() != nil {
		log.Printf("IPv6 not supported yet")
		return nil, ErrProtocolNotSupported
	} else {
		log.Printf("Unknown IP version")
		return nil, ErrProtocolNotSupported
	}
}

func (neigh *neighborSubsystem) SendRequest(ip net.IP, iface netstack.NetworkInterface) {
	if ip.To4() != nil {
		neigh.arp.ARPRequest(ip, iface)
	} else if ip.To16() != nil {
		log.Printf("IPv6 not supported yet")
	} else {
		log.Printf("Unknown IP version")
	}
}
