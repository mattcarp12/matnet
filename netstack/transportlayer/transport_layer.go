package transportlayer

import (
	logging "log"
	"os"

	"github.com/mattcarp12/go-net/netstack"
	"github.com/mattcarp12/go-net/netstack/networklayer"
)

var log = logging.New(os.Stdout, "[Transport Layer] ", logging.Ldate|logging.Lmicroseconds|logging.Lshortfile)

type TransportLayer struct {
	netstack.ILayer
}

func Init(network_layer *networklayer.NetworkLayer) *TransportLayer {
	transport_layer := &TransportLayer{}
	transport_layer.SkBuffReaderWriter = netstack.NewSkBuffChannels()

	// Create Transport Layer protocols
	tcp := NewTCP()
	udp := NewUDP()

	// Add Transport Layer protocols to Transport Layer
	transport_layer.AddProtocol(tcp)
	transport_layer.AddProtocol(udp)

	// Set Transport Layer as the Layer for the protocols
	tcp.SetLayer(transport_layer)
	udp.SetLayer(transport_layer)

	// Set Transport Layer as the next layer for Network Layer
	network_layer.SetNextLayer(transport_layer)

	// Set Network Layer as previous layer for Transport Layer
	transport_layer.SetPrevLayer(network_layer)

	// Start protocol goroutines
	netstack.StartProtocol(tcp)
	netstack.StartProtocol(udp)

	// Start transport layer goroutines
	netstack.StartLayerDispatchLoops(transport_layer)

	return transport_layer
}

func set_skb_type(skb *netstack.SkBuff) error {
	// Check the address type of the destination address
	// If it's IPv4, set the skb type to IPv4
	// If it's IPv6, set the skb type to IPv6
	// If it's neither, error
	addrType := skb.GetSocket().GetDestinationAddress().GetType()
	if addrType == netstack.AddressTypeIPv4 {
		skb.SetType(netstack.ProtocolTypeIPv4)
	} else if addrType == netstack.AddressTypeIPv6 {
		skb.SetType(netstack.ProtocolTypeIPv6)
	} else {
		return netstack.ErrInvalidAddressType
	}
	return nil
}
