package transportlayer

import (
	"github.com/mattcarp12/go-net/netstack"
	"github.com/mattcarp12/go-net/networklayer"
	"github.com/mattcarp12/go-net/transportlayer/tcp"
	"github.com/mattcarp12/go-net/transportlayer/udp"
)

type TransportLayer struct {
	netstack.ILayer
}

func Init(network_layer *networklayer.NetworkLayer) *TransportLayer {
	transport_layer := &TransportLayer{}
	transport_layer.SkBuffReaderWriter = netstack.NewSkBuffChannels()

	// Create Transport Layer protocols
	tcp := tcp.NewTCP()
	udp := udp.NewUDP()

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
