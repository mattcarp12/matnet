package linklayer

import (
	"errors"
	"log"
	"net"

	"github.com/mattcarp12/matnet/netstack"
)

type NeighborProtocol interface {
	netstack.Protocol
	Resolve(ip net.IP) (net.HardwareAddr, error)
	SendRequest(skb *netstack.SkBuff)
}

type NeighborSubsystem struct {
	protocols map[netstack.ProtocolType]NeighborProtocol
	Log       *log.Logger
}

func NewNeighborSubsystem() *NeighborSubsystem {
	return &NeighborSubsystem{
		protocols: make(map[netstack.ProtocolType]NeighborProtocol),
		Log:       netstack.NewLogger("NeighborSubsystem"),
	}
}

func (neigh *NeighborSubsystem) AddProtocol(protocol NeighborProtocol) {
	neigh.protocols[protocol.GetType()] = protocol
}

var ErrProtocolNotSupported = errors.New("protocol not supported")

func (neigh *NeighborSubsystem) Resolve(ip net.IP) (net.HardwareAddr, error) {
	if ip.To4() != nil {
		return neigh.protocols[netstack.ProtocolTypeARP].Resolve(ip)
	} else if ip.To16() != nil {
		return nil, ErrProtocolNotSupported
	} else {
		return nil, ErrProtocolNotSupported
	}
}

// SendRequest sends an ARP request to the specified IP address
// This function should be called as a goroutine
func (neigh *NeighborSubsystem) SendRequest(skb *netstack.SkBuff) {
	dstIP := skb.GetDstIP()
	if dstIP == nil {
		return
	}

	if dstIP.To4() != nil {
		neigh.protocols[netstack.ProtocolTypeARP].SendRequest(skb)
	} else if dstIP.To16() != nil {
		neigh.Log.Printf("IPv6 not supported yet")
	} else {
		neigh.Log.Printf("Unknown IP version")
	}
}

func (neigh *NeighborSubsystem) HandleRx(skb *netstack.SkBuff) {
	// Dispatch the request to the appropriate protocol
	protocol := neigh.protocols[skb.GetType()]
	if protocol != nil {
		protocol.HandleRx(skb)
	} else {
		neigh.Log.Printf("Protocol not supported")
	}
}
