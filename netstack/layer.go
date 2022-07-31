package netstack

// ===========================================================================
// Layer represents a layer in the network stack (e.g. link layer, network layer, transport layer)
// A layer consists of a set of protocols.
// A layer is responsible for dispatching SkBuffs to the appropriate protocol.
// A layer also contains pointers to its neighboring layers.
// ===========================================================================
type Layer struct {
	SkBuffReaderWriter
	protocols map[ProtocolType]Protocol
	nextLayer *Layer
	prevLayer *Layer
}

func NewLayer(protocols ...Protocol) *Layer {
	protocolMap := make(map[ProtocolType]Protocol)
	for _, protocol := range protocols {
		protocolMap[protocol.GetType()] = protocol
	}

	return &Layer{
		protocols:          protocolMap,
		SkBuffReaderWriter: NewSkBuffChannels(),
	}
}

func (layer Layer) GetProtocol(protocolType ProtocolType) (Protocol, error) {
	protocol, ok := layer.protocols[protocolType]
	if !ok {
		return nil, ErrProtocolNotFound
	}

	return protocol, nil
}

func (layer Layer) GetNextLayer() *Layer {
	return layer.nextLayer
}

func (layer *Layer) SetNextLayer(nextLayer *Layer) {
	layer.nextLayer = nextLayer
}

func (layer Layer) GetPrevLayer() *Layer {
	return layer.prevLayer
}

func (layer *Layer) SetPrevLayer(prevLayer *Layer) {
	layer.prevLayer = prevLayer
}

/*
	A layer has two goroutines that run concurrently:
		1. RxDispatch - reads SkBuffs from the layer's rx_chan and dispatches them to the appropriate protocol
		2. TxDispatch - reads SkBuffs from the layer's tx_chan and dispatches them to the appropriate protocol
*/

func (layer Layer) StartLayer() {
	go layer.RxDispatch()
	go layer.TxDispatch()
}

func (layer Layer) RxDispatch() {
	for {
		// Layer reads SkBuff from it's rx_chan
		skb := <-layer.RxChan()

		// Dispatch skb to appropriate protocol
		protocol, err := layer.GetProtocol(skb.GetType())
		if err != nil {
			continue
		}

		// Send skb to protocol
		protocol.RxChan() <- skb
	}
}

func (layer Layer) TxDispatch() {
	for {
		// Layer reads skb from it's tx_chan
		skb := <-layer.TxChan()

		// Dispatch skb to appropriate protocol
		protocol, err := layer.GetProtocol(skb.GetType())
		if err != nil {
			continue
		}

		// Send skb to protocol
		protocol.TxChan() <- skb
	}
}
