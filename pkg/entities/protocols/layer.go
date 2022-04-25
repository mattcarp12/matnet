package protocols

import "log"

// Layer represents a layer in the network stack
// A layer consists of a set of protocols
// A layer is responsible for dispatching PDUs to the appropriate protocol
// The protocol processes PDUs and sends them to the next layer (up or down)
type Layer interface {
	GetProtocol(ProtocolType) (Protocol, error)
	AddProtocol(ProtocolType, Protocol)
	PDUReaderWriter

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

func (layer *ILayer) AddProtocol(protocolType ProtocolType, protocol Protocol) {
	if layer.protocols == nil {
		layer.protocols = make(map[ProtocolType]Protocol)
	}
	layer.protocols[protocolType] = protocol
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
		// Layer reads PDU from it's rx_chan
		pdu := <-layer.RxChan()

		// Dispatch PDU to appropriate protocol
		protocol, err := layer.GetProtocol(pdu.GetType())
		if err != nil {
			log.Printf("Error getting protocol: %v", err)
			continue
		}

		// Handle PDU
		protocol.HandleRx(pdu)

	}
}

func TxDispatch(layer Layer) {
	for {
		// Layer reads PDU from it's tx_chan
		pdu := <-layer.TxChan()

		// Dispatch PDU to appropriate protocol
		protocol, err := layer.GetProtocol(pdu.GetType())
		if err != nil {
			log.Printf("Error getting protocol: %v", err)
			continue
		}

		// Handle PDU
		protocol.HandleTx(pdu)

	}
}
