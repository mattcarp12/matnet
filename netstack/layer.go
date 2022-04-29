package netstack

import "log"

// Layer represents a layer in the network stack
// A layer consists of a set of protocols
// A layer is responsible for dispatching SkBuffs to the appropriate protocol
// The protocol processes SkBuffs and sends them to the next layer (up or down)
type Layer interface {
	GetProtocol(ProtocolType) (Protocol, error)
	AddProtocol(Protocol)
	SkBuffReaderWriter

	GetNextLayer() Layer
	SetNextLayer(Layer)
	GetPrevLayer() Layer
	SetPrevLayer(Layer)
}

// ILayer is an implementation of Layer interface
// to be used by other layer implementations
type ILayer struct {
	protocols map[ProtocolType]Protocol
	nextLayer Layer
	prevLayer Layer
}

func (layer *ILayer) GetProtocol(protocolType ProtocolType) (Protocol, error) {
	protocol, ok := layer.protocols[protocolType]
	if !ok {
		return nil, ErrProtocolNotFound
	}
	return protocol, nil
}

func (layer *ILayer) AddProtocol(protocol Protocol) {
	if layer.protocols == nil {
		layer.protocols = make(map[ProtocolType]Protocol)
	}
	layer.protocols[protocol.GetType()] = protocol
}

func (layer *ILayer) GetNextLayer() Layer {
	return layer.nextLayer
}

func (layer *ILayer) SetNextLayer(nextLayer Layer) {
	layer.nextLayer = nextLayer
}

func (layer *ILayer) GetPrevLayer() Layer {
	return layer.prevLayer
}

func (layer *ILayer) SetPrevLayer(prevLayer Layer) {
	layer.prevLayer = prevLayer
}

func RxDispatch(layer Layer) {
	for {
		// Layer reads SkBuff from it's rx_chan
		skb := <-layer.RxChan()

		// Dispatch skb to appropriate protocol
		protocol, err := layer.GetProtocol(skb.GetType())
		if err != nil {
			log.Printf("Error getting protocol: %v", err)
			continue
		}

		// Handle skb
		protocol.HandleRx(skb)

	}
}

func TxDispatch(layer Layer) {
	for {
		// Layer reads skb from it's tx_chan
		skb := <-layer.TxChan()

		// Dispatch skb to appropriate protocol
		protocol, err := layer.GetProtocol(skb.GetType())
		if err != nil {
			log.Printf("Error getting protocol: %v", err)
			continue
		}

		// Handle skb
		protocol.HandleTx(skb)

	}
}
