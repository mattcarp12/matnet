package netstack

// Protocol represents a single network protocol
// A protocol is responsible for receiving packets from the
// layer it belongs to, and processing the packets per it's
// individual protocol logic.
type Protocol interface {
	GetType() ProtocolType
	GetLayer() Layer
	SetLayer(Layer)
	SkBuffReaderWriter
	HandleRx(*SkBuff)
	HandleTx(*SkBuff)
}

// IProtocol used by other protocols to implement Protocol interface.
type IProtocol struct {
	// Every protocol has rx and tx channels
	SkBuffChannels

	// ProtocolType is the type of the protocol
	protocolType ProtocolType

	// The layer this protocol belongs to
	layer Layer
}

func NewIProtocol(protocolType ProtocolType) IProtocol {
	return IProtocol{
		protocolType:   protocolType,
		SkBuffChannels: NewSkBuffChannels(),
	}
}

func (protocol *IProtocol) GetType() ProtocolType {
	return protocol.protocolType
}

func (protocol *IProtocol) GetLayer() Layer {
	return protocol.layer
}

func (protocol *IProtocol) SetLayer(layer Layer) {
	protocol.layer = layer
}

/*
	ProtocolXXLoop used to start the Rx and Tx loops for each protocol.
*/

func ProtocolLoop(protocol Protocol) {
	go ProtocolRxLoop(protocol)
	go ProtocolTxLoop(protocol)
}

func ProtocolRxLoop(protocol Protocol) {
	for {
		// Network protocol reads skb from it's rx_chan
		skb := <-protocol.RxChan()

		// Handle sk_buff
		protocol.HandleRx(skb)
	}
}

func ProtocolTxLoop(protocol Protocol) {
	for {
		// Network protocol reads skb from it's tx_chan
		skb := <-protocol.TxChan()

		// Handle sk_buff
		protocol.HandleTx(skb)
	}
}
