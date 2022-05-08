package netstack

import (
	"errors"
	"net"
	"strconv"

	"github.com/google/uuid"
)

/*
	Socket objects represent a network endpoint that can be connected to
	or listened on.
*/

// Individial socket type implementations must define their own SocketOperations methods
type SocketOperations interface {
	Bind(addr SockAddr) error
	Listen(backlog int) error
	Accept() (net.Conn, error)
	Connect(addr SockAddr) error
	Close() error
	Read(b []byte) (int, error)
	Write(b []byte) (int, error)
	ReadFrom(b []byte, addr *SockAddr) (int, error)
	WriteTo(b []byte, addr SockAddr) (int, error)
}

type SocketMetaOps interface {
	GetType() SocketType
	SetType(t SocketType)

	GetProtocol() Protocol
	SetProtocol(p Protocol)

	GetSourceAddress() SockAddr
	SetSourceAddress(addr SockAddr)

	GetDestinationAddress() SockAddr
	SetDestinationAddress(addr SockAddr)

	GetState() SocketState
	SetState(SocketState)

	GetID() SockID
	SetID(id SockID)
}

type SocketSkbOps interface {
	RecvSkb() *SkBuff
	SendSkb(skb *SkBuff)
}

type Socket interface {
	SocketOperations
	SocketMetaOps
	SocketSkbOps
}

type SockID string

var ErrInvalidSocketID = errors.New("Invalid socket ID")

func NewSockID() SockID {
	return SockID(uuid.New().String())
}

type SocketType uint

const (
	SocketTypeInvalid SocketType = iota
	SocketTypeStream
	SocketTypeDatagram
	SocketTypeRaw
)

var ErrInvalidSocketType = errors.New("Invalid socket type")

type SocketState uint

const (
	SocketStateInvalid SocketState = iota
	SocketStateUnbound
	SocketStateBound
	SocketStateConnected
	SocketStateListening
	SocketStateClosed
)

/*
	SockAddr -- generic structure for network addresses
	Can hold either an IPv4 or IPv6 address
*/
type SockAddr struct {
	Port uint16
	IP   net.IP
}

type AddressType uint

const (
	AddressTypeIPv4 = iota
	AddressTypeIPv6
	AddressTypeUnknown
)

var ErrInvalidAddressType = errors.New("Invalid address type")

func (s SockAddr) GetPort() uint16 {
	return s.Port
}

func (s SockAddr) GetIP() net.IP {
	return s.IP
}

func (s SockAddr) String() string {
	return s.IP.String() + ":" + strconv.Itoa(int(s.Port))
}

func (s SockAddr) GetType() AddressType {
	if s.IP.To4() != nil {
		return AddressTypeIPv4
	} else if s.IP.To16() != nil {
		return AddressTypeIPv6
	} else {
		return AddressTypeUnknown
	}
}

/*
	ISocket - helper struct for Socket implementations
*/

type ISocket struct {
	// Socket type
	Type SocketType

	// Socket protocol
	Protocol Protocol

	// Source address
	SrcAddr SockAddr

	// Destination address
	DestAddr SockAddr

	// Socket state
	State SocketState

	// Socket ID
	ID SockID

	// Routing table
	RoutingTable RoutingTable
}

func NewSocket() *ISocket {
	return &ISocket{}
}

func (s *ISocket) GetType() SocketType {
	return s.Type
}

func (s *ISocket) SetType(t SocketType) {
	s.Type = t
}

func (s *ISocket) GetProtocol() Protocol {
	return s.Protocol
}

func (s *ISocket) SetProtocol(p Protocol) {
	s.Protocol = p
}

func (s *ISocket) GetSourceAddress() SockAddr {
	return s.SrcAddr
}

func (s *ISocket) SetSourceAddress(addr SockAddr) {
	s.SrcAddr = addr
}

func (s *ISocket) GetDestinationAddress() SockAddr {
	return s.DestAddr
}

func (s *ISocket) SetDestinationAddress(addr SockAddr) {
	s.DestAddr = addr
}

func (s *ISocket) GetState() SocketState {
	return s.State
}

func (s *ISocket) SetState(state SocketState) {
	s.State = state
}

func (s *ISocket) GetID() SockID {
	return s.ID
}

func (s *ISocket) SetID(id SockID) {
	s.ID = id
}

func (s *ISocket) GetRoutingTable() RoutingTable {
	return s.RoutingTable
}

func (s *ISocket) SetRoutingTable(rt RoutingTable) {
	s.RoutingTable = rt
}

func (s *ISocket) RecvSkb() *SkBuff {
	return <-s.GetProtocol().RxChan()
}

func (s *ISocket) SendSkb(skb *SkBuff) {
	s.GetProtocol().TxChan() <- skb
}

/*
	Socket Manager - Interface between the network stack and the application
*/

type SocketManager interface {
	// Create a new socket
	CreateSocket(SocketType) (Socket, error)

	// Get a socket by ID
	GetSocket(SockID) (Socket, error)
}
