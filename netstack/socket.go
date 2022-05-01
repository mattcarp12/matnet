package netstack

import "net"

/*
	Socket objects represent a network endpoint that can be connected to
	or listened on.
*/

type Socket interface {
	// Get the socket type
	GetType() SocketType

	// Get the socket protocol
	GetProtocol() Protocol

	// Get the socket address
	GetAddress() SocketAddress

	// Get the socket state
	GetState() SocketState

	// Set the socket state
	SetState(SocketState)
}

type SocketType uint

const (
	SocketTypeInvalid SocketType = iota
	SocketTypeStream
	SocketTypeDatagram
	SocketTypeRaw
	SocketTypeRdm
	SocketTypeSeqPacket
)

type SocketState uint

const (
	SocketStateInvalid SocketState = iota
	SocketStateUnbound
	SocketStateBound
	SocketStateConnected
	SocketStateListening
	SocketStateClosed
)

type SocketAddress struct {
	// Source port
	SrcPort uint16

	// Destination port
	DestPort uint16

	// Source IP
	SrcIP net.IP

	// Destination IP
	DestIP net.IP
}

var _ Socket = &socket{}

type socket struct {
	// Socket type
	Type SocketType

	// Socket protocol
	Protocol Protocol

	// Socket address
	Address SocketAddress

	// Socket state
	State SocketState
}

func NewSocket() *socket {
	return &socket{
		Type:     SocketTypeInvalid,
		Protocol: nil,
		Address: SocketAddress{
			SrcPort:  0,
			DestPort: 0,
			SrcIP:    nil,
			DestIP:   nil,
		},
		State: SocketStateInvalid,
	}
}

func (s *socket) GetType() SocketType {
	return s.Type
}

func (s *socket) GetProtocol() Protocol {
	return s.Protocol
}

func (s *socket) GetAddress() SocketAddress {
	return s.Address
}

func (s *socket) GetState() SocketState {
	return s.State
}

func (s *socket) SetState(state SocketState) {
	s.State = state
}
