package netstack

import "errors"

/*
	Protocol represents a single network protocol
	A protocol is responsible for receiving packets from the
	layer it belongs to, and processing the packets per it's
	individual protocol logic.

	Protocols receive packets from their RxChan, and handle
	them in their HandleRx function. Here is the general flow:
		1. HandleRx is called with a skb
		2. Create a new header object
		3. Unmarshal the header object, handle errors
		4. Check the headers destination address matches the
			protocol's address
		5. Set the header object in the skb
		6. Strip the header from the skb
		7. Set the skb type to the protocol type of the next protocol
			(e.g. the EtherType for L2 or Protocol for L3)
		8. Send the skb to the next layer using RxUp()
*/

type ProtocolType uint16

const (
	ProtocolTypeEthernet ProtocolType = iota
	ProtocolTypeIPv4
	ProtocolTypeICMPv4
	ProtocolTypeARP
	ProtocolTypeIPv6
	ProtocolTypeICMPv6
	ProtocolTypeTCP
	ProtocolTypeUDP
	ProtocolTypeRaw
	ProtocolTypeUnknown ProtocolType = 0xFFFF
)

var ErrProtocolNotFound = errors.New("protocol not found")

type Protocol interface {
	GetType() ProtocolType
	GetLayer() Layer
	SetLayer(Layer)
	SkBuffReaderWriter
	HandleRx(*SkBuff)
	HandleTx(*SkBuff)
	RxUp(*SkBuff)   // Used to send packets to the next layer up the stack
	TxDown(*SkBuff) // Used to send packets to the next layer down the stack
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

func (protocol *IProtocol) RxUp(skb *SkBuff) {
	// Send skb to next layer up the stack
	protocol.layer.GetNextLayer().RxChan() <- skb
}

func (protocol *IProtocol) TxDown(skb *SkBuff) {
	// Send skb to next layer down the stack
	protocol.layer.GetPrevLayer().TxChan() <- skb
}

/*
	ProtocolXXLoop used to start the Rx and Tx loops for each protocol.
*/

func StartProtocol(protocol Protocol) {
	go ProtocolRxLoop(protocol)
	go ProtocolTxLoop(protocol)
}

func ProtocolRxLoop(protocol Protocol) {
	for {
		// Network protocol reads skb from it's rx_chan
		skb := <-protocol.RxChan()

		// Handle skb
		protocol.HandleRx(skb)
	}
}

func ProtocolTxLoop(protocol Protocol) {
	for {
		// Network protocol reads skb from it's tx_chan
		skb := <-protocol.TxChan()

		// Handle skb
		protocol.HandleTx(skb)
	}
}
