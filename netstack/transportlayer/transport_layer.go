package transportlayer

import (
	"errors"

	"github.com/mattcarp12/matnet/netstack"
	"github.com/mattcarp12/matnet/netstack/networklayer"
)

type TransportLayer struct {
	netstack.ILayer
}

func Init(networkLayer *networklayer.NetworkLayer) *TransportLayer {
	transportLayer := &TransportLayer{}
	transportLayer.SkBuffReaderWriter = netstack.NewSkBuffChannels()

	// Create Transport Layer protocols
	tcp := NewTCP()
	udp := NewUDP()

	// Add Transport Layer protocols to Transport Layer
	transportLayer.AddProtocol(tcp)
	transportLayer.AddProtocol(udp)

	// Set Transport Layer as the Layer for the protocols
	tcp.SetLayer(transportLayer)
	udp.SetLayer(transportLayer)

	// Set Transport Layer as the next layer for Network Layer
	networkLayer.SetNextLayer(transportLayer)

	// Set Network Layer as previous layer for Transport Layer
	transportLayer.SetPrevLayer(networkLayer)

	// Start protocol goroutines
	netstack.StartProtocol(tcp)
	netstack.StartProtocol(udp)

	// Start transport layer goroutines
	netstack.StartLayer(transportLayer)

	return transportLayer
}

func setSkbType(skb *netstack.SkBuff) error {
	// Check the address type of the destination address
	// If it's IPv4, set the skb type to IPv4
	// If it's IPv6, set the skb type to IPv6
	// If it's neither, error
	dstAddr := skb.GetDstIP()
	if dstAddr == nil {
		return errors.New("destination address is nil")
	}
	ip4 := dstAddr.To4()
	ip6 := dstAddr.To16()
	if ip4 != nil {
		skb.SetType(netstack.ProtocolTypeIPv4)
	} else if ip6 != nil {
		skb.SetType(netstack.ProtocolTypeIPv6)
	} else {
		return errors.New("destination address is neither IPv4 nor IPv6")
	}
	return nil
}

func ConnectionID(localAddr netstack.SockAddr, remoteAddr netstack.SockAddr) string {
	return localAddr.String() + "-" + remoteAddr.String()
}
